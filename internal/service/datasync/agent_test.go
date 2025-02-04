package datasync_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDataSyncAgent_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var agent1 datasync.DescribeAgentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`agent/agent-.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", ""),
					resource.TestCheckResourceAttr(resourceName, "private_link_endpoint", ""),
					resource.TestCheckResourceAttr(resourceName, "security_group_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "ip_address"},
			},
		},
	})
}

func TestAccDataSyncAgent_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var agent1 datasync.DescribeAgentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceAgent(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncAgent_agentName(t *testing.T) {
	ctx := acctest.Context(t)
	var agent1, agent2 datasync.DescribeAgentOutput
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_name(rName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccAgentConfig_name(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "ip_address"},
			},
		},
	})
}

func TestAccDataSyncAgent_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var agent1, agent2, agent3 datasync.DescribeAgentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "ip_address"},
			},
			{
				Config: testAccAgentConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent2),
					testAccCheckAgentNotRecreated(&agent1, &agent2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAgentConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent3),
					testAccCheckAgentNotRecreated(&agent2, &agent3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccDataSyncAgent_vpcEndpointID(t *testing.T) {
	ctx := acctest.Context(t)
	var agent datasync.DescribeAgentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"
	securityGroupResourceName := "aws_security_group.test"
	subnetResourceName := "aws_subnet.test.0"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_vpcEndpointID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, resourceName, &agent),
					resource.TestCheckResourceAttr(resourceName, "security_group_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_arns.*", securityGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_arns.*", subnetResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_endpoint_id", vpcEndpointResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "ip_address", "private_link_ip"},
			},
		},
	})
}

func testAccCheckAgentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_agent" {
				continue
			}

			_, err := tfdatasync.FindAgentByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Agent %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAgentExists(ctx context.Context, resourceName string, agent *datasync.DescribeAgentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn(ctx)

		output, err := tfdatasync.FindAgentByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*agent = *output

		return nil
	}
}

func testAccCheckAgentNotRecreated(i, j *datasync.DescribeAgentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("DataSync Agent was recreated")
		}

		return nil
	}
}

func testAccAgentAgentConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/datasync/latest/userguide/deploy-agents.html
data "aws_ssm_parameter" "aws_service_datasync_ami" {
  name = "/aws/service/datasync/ami"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test[0].id
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  depends_on = [aws_internet_gateway.test]

  ami                         = data.aws_ssm_parameter.aws_service_datasync_ami.value
  associate_public_ip_address = true

  # Default instance type from sync.sh
  instance_type          = "c5.2xlarge"
  vpc_security_group_ids = [aws_security_group.test.id]
  subnet_id              = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAgentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_base(rName), `
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
}
`)
}

func testAccAgentConfig_name(rName, agentName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}
`, agentName))
}

func testAccAgentConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccAgentConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2))
}

func testAccAgentConfig_vpcEndpointID(rName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  name                  = %[1]q
  security_group_arns   = [aws_security_group.test.arn]
  subnet_arns           = [aws_subnet.test[0].arn]
  vpc_endpoint_id       = aws_vpc_endpoint.test.id
  ip_address            = aws_instance.test.public_ip
  private_link_endpoint = data.aws_network_interface.test.private_ip
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  service_name       = "com.amazonaws.${data.aws_region.current.name}.datasync"
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = [aws_subnet.test[0].id]
  vpc_endpoint_type  = "Interface"

  tags = {
    Name = %[1]q
  }
}

data "aws_network_interface" "test" {
  id = tolist(aws_vpc_endpoint.test.network_interface_ids)[0]
}
`, rName))
}
