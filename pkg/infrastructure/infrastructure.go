package infrastructure

import (
	"context"
	"fmt"
	"io"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/zchase/jacuik/pkg/jacuik_config"
	ec2x "github.com/zchase/pulumi-awsx-go/sdk/go/awsx-go/ec2"
	ecrx "github.com/zchase/pulumi-awsx-go/sdk/go/awsx-go/ecr"
	ecsx "github.com/zchase/pulumi-awsx-go/sdk/go/awsx-go/ecs"
	lbx "github.com/zchase/pulumi-awsx-go/sdk/go/awsx-go/lb"
)

func deployInfrastructure(name string, config *jacuik_config.AppConfig) func(ctx *pulumi.Context) error {
	return func(ctx *pulumi.Context) error {
		// Create a VPC
		vpcName := fmt.Sprintf("%s-vpc", name)
		vpc, err := ec2x.NewVpc(ctx, vpcName, nil)
		if err != nil {
			return err
		}

		// Create the cluster
		clusterName := fmt.Sprintf("%s-cluster", name)
		cluster, err := ecs.NewCluster(ctx, clusterName, nil)
		if err != nil {
			return err
		}

		albName := fmt.Sprintf("%s-alb", name)
		alb, err := lbx.NewApplicationLoadBalancer(ctx, albName, &lbx.ApplicationLoadBalancerArgs{
			SubnetIds: vpc.PublicSubnetIds,
		})
		if err != nil {
			return err
		}

		publicListener := alb.Listeners.ApplyT(func(listeners []*lb.Listener) lb.Listener {
			return *listeners[0]
		}).(pulumi.AnyOutput)

		repositoryName := fmt.Sprintf("%s-repository", name)
		repository, err := ecrx.NewRepository(ctx, repositoryName, nil)
		if err != nil {
			return err
		}

		for _, svc := range config.Services {
			imageName := fmt.Sprintf("%s-%s-image", name, svc.Name)
			image, err := ecrx.NewImage(ctx, imageName, &ecrx.ImageArgs{
				RepositoryUrl: repository.Url,
				Path:          pulumi.String(svc.PathToDockerfile),
			})
			if err != nil {
				return err
			}

			defs := []ecsx.TaskDefinitionPortMappingInput{
				ecsx.TaskDefinitionPortMappingArgs{
					ContainerPort: publicListener.ApplyT(func(x interface{}) pulumi.IntPtrOutput {
						l := x.(lb.Listener)
						return l.Port
					}).(pulumi.IntPtrOutput),
					TargetGroup: alb.DefaultTargetGroup,
				},
			}

			cloudSvcName := fmt.Sprintf("%s-%s-svc", name, svc.Name)
			_, err = ecsx.NewFargateService(ctx, cloudSvcName, &ecsx.FargateServiceArgs{
				Cluster:      cluster.Arn,
				DesiredCount: pulumi.IntPtr(1),
				NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
					Subnets:        vpc.PublicSubnetIds,
					AssignPublicIp: pulumi.BoolPtr(true),
					SecurityGroups: alb.DefaultSecurityGroup.ApplyT(func(sg *ec2.SecurityGroup) pulumi.StringArrayOutput {
						result := []pulumi.StringOutput{sg.ID().ToStringOutput()}
						return pulumi.ToStringArrayOutput(result)
					}).(pulumi.StringArrayOutput),
				},
				TaskDefinitionArgs: &ecsx.FargateServiceTaskDefinitionArgs{
					Container: &ecsx.TaskDefinitionContainerDefinitionArgs{
						Image:        image.ImageUri,
						Cpu:          pulumi.IntPtr(102),
						Memory:       pulumi.IntPtr(50),
						PortMappings: ecsx.TaskDefinitionPortMappingArray(defs),
					},
				},
			})
			if err != nil {
				return err
			}
		}

		ctx.Export("serviceUrl", alb.LoadBalancer.DnsName())
		return nil
	}
}

type InfrastructureHandler struct {
	Name   string
	Config *jacuik_config.AppConfig
}

func NewInfrastructureHandler(name string, config *jacuik_config.AppConfig) *InfrastructureHandler {
	return &InfrastructureHandler{
		Name:   name,
		Config: config,
	}
}

func (i *InfrastructureHandler) Preview(progressWriter io.Writer) error {
	ctx, stack, err := i.configureApplicationStack()
	if err != nil {
		return err
	}

	stdoutStreamer := optpreview.ProgressStreams(progressWriter)

	_, err = stack.Preview(ctx, stdoutStreamer)
	if err != nil {
		return err
	}

	return nil
}

func (i *InfrastructureHandler) Update(progressWriter io.Writer) error {
	ctx, stack, err := i.configureApplicationStack()
	if err != nil {
		return err
	}

	stdoutStreamer := optup.ProgressStreams(progressWriter)

	_, err = stack.Up(ctx, stdoutStreamer)
	if err != nil {
		return err
	}

	return nil
}

func (i *InfrastructureHandler) Destroy(progressWriter io.Writer) error {
	ctx, stack, err := i.configureApplicationStack()
	if err != nil {
		return err
	}

	stdoutStreamer := optdestroy.ProgressStreams(progressWriter)

	_, err = stack.Destroy(ctx, stdoutStreamer)
	if err != nil {
		return err
	}

	return nil
}

func (i *InfrastructureHandler) configureApplicationStack() (context.Context, auto.Stack, error) {
	ctx := context.Background()

	// TODO: make this configurable via environments.
	projectName := i.Name
	stackName := "dev"

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, deployInfrastructure(i.Name, i.Config))
	if err != nil {
		return ctx, auto.Stack{}, err
	}

	workspace := stack.Workspace()

	// Plugins
	err = workspace.InstallPlugin(ctx, "aws", "v5.6.0")
	if err != nil {
		return ctx, auto.Stack{}, err
	}

	// TODO: enable this when awsx-go is available.
	// err = workspace.InstallPlugin(ctx, "awsx-go", "v0.0.1")
	// if err != nil {
	// 	return ctx, auto.Stack{}, err
	// }

	// TODO: make this configurable
	err = stack.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: "us-west-2"})
	if err != nil {
		return ctx, auto.Stack{}, err
	}

	return ctx, stack, nil
}
