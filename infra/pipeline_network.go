package main

import (
	awsec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	elbv2 "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewNetwork creates a public ALB where prod (80) is public and test (9001) is restricted.
// Pass the allowed CIDR ranges for test access via testAllowedCidrs.
func NewNetwork(
	scope constructs.Construct,
	id string,
	testAllowedCidrs []*string,
) (
	awsec2.IVpc,
	elbv2.ApplicationLoadBalancer,
	elbv2.ApplicationListener,
	elbv2.ApplicationListener,
	elbv2.ApplicationTargetGroup,
	elbv2.ApplicationTargetGroup,
) {
	vpc := awsec2.NewVpc(scope, jsii.String(id+"Vpc"), &awsec2.VpcProps{
		MaxAzs: jsii.Number(2),
	})

	// Security group for the ALB: open 80 to the world, restrict 9001 to provided CIDRs
	albSg := awsec2.NewSecurityGroup(scope, jsii.String(id+"AlbSg"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description:      jsii.String("ALB SG: public prod (80), restricted test (9001)"),
	})
	// Prod open to the internet
	albSg.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(jsii.Number(80)),
		jsii.String("Allow HTTP to prod listener"),
		nil, // remoteRule
	)
	// Test only from trusted CIDRs
	for _, cidr := range testAllowedCidrs {
		albSg.AddIngressRule(
			awsec2.Peer_Ipv4(cidr),
			awsec2.Port_Tcp(jsii.Number(9001)),
			jsii.String("Allow test listener from trusted CIDRs"),
			nil, // remoteRule
		)
	}

	alb := elbv2.NewApplicationLoadBalancer(scope, jsii.String(id+"Alb"),
		&elbv2.ApplicationLoadBalancerProps{
			Vpc:            vpc,
			InternetFacing: jsii.Bool(true),
			SecurityGroup:  albSg,
		})

	// Do not auto-open; SG above explicitly controls ingress
	prod := alb.AddListener(jsii.String("ProdListener"),
		&elbv2.BaseApplicationListenerProps{
			Port:     jsii.Number(80),
			Protocol: elbv2.ApplicationProtocol_HTTP,
			Open:     jsii.Bool(false),
		})
	test := alb.AddListener(jsii.String("TestListener"),
		&elbv2.BaseApplicationListenerProps{
			Port:     jsii.Number(9001),
			Protocol: elbv2.ApplicationProtocol_HTTP,
			Open:     jsii.Bool(false),
		})

	blueTG := elbv2.NewApplicationTargetGroup(scope, jsii.String("BlueTG"),
		&elbv2.ApplicationTargetGroupProps{
			Vpc:        vpc,
			Port:       jsii.Number(8888),
			Protocol:   elbv2.ApplicationProtocol_HTTP,
			TargetType: elbv2.TargetType_IP,
			HealthCheck: &elbv2.HealthCheck{
				Path:             jsii.String("/health"),
				HealthyHttpCodes: jsii.String("200-399"),
			},
		})
	greenTG := elbv2.NewApplicationTargetGroup(scope, jsii.String("GreenTG"),
		&elbv2.ApplicationTargetGroupProps{
			Vpc:        vpc,
			Port:       jsii.Number(8888),
			Protocol:   elbv2.ApplicationProtocol_HTTP,
			TargetType: elbv2.TargetType_IP,
			HealthCheck: &elbv2.HealthCheck{
				Path:             jsii.String("/health"),
				HealthyHttpCodes: jsii.String("200-399"),
			},
		})

	prod.AddTargetGroups(jsii.String("ProdTG"),
		&elbv2.AddApplicationTargetGroupsProps{
			TargetGroups: &[]elbv2.IApplicationTargetGroup{blueTG},
		})
	test.AddTargetGroups(jsii.String("TestTG"),
		&elbv2.AddApplicationTargetGroupsProps{
			TargetGroups: &[]elbv2.IApplicationTargetGroup{greenTG},
		})

	return vpc, alb, prod, test, blueTG, greenTG
}
