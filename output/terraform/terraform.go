package terraform

import (
	"bytes"
	"html/template"
	"net"
)

func Output(vpccidr *net.IPNet, alloc []map[string]*net.IPNet) (string, error) {
	var buf bytes.Buffer
	tmplPreamble, err := template.New("preamble").Parse(`variable "cidr_block" {
    type = "string"
    default = "{{.cidrblock}}"
}

# Specify the provider and access details
provider "aws" {

}

data "aws_region" "default" {
  current = true
}

# Create a VPC to launch our instances into
resource "aws_vpc" "default" {
    cidr_block = "${var.cidr_block}"
    enable_dns_hostnames = true
}

# Grant the VPC internet access on its main route table
resource "aws_route" "internet_access" {
    route_table_id         = "${aws_vpc.default.main_route_table_id}"
    destination_cidr_block = "0.0.0.0/0"
    gateway_id             = "${aws_internet_gateway.default.id}"
}

resource "aws_internet_gateway" "default" {
  vpc_id = "${aws_vpc.default.id}"

    tags {
            Name = "vpc-igw"
          }
}`)

	if err != nil {
		return "", err
	}

	tmplAZ, err := template.New("az").Parse(`

resource "aws_subnet" "AZ-{{.az}}-{{.function}}" {
vpc_id                  = "${aws_vpc.default.id}"
cidr_block              = "{{.cidrblockInner}}"
availability_zone       = "${data.aws_region.default.name}{{.az}}"
map_public_ip_on_launch = false
}`)

	if err != nil {
		return "", err
	}

	tmplPreamble.Execute(&buf, map[string]string{"cidrblock": vpccidr.String()})
	for k, v := range alloc {
		for _, t := range []string{"public", "private", "protected"} {
			tmplAZ.Execute(&buf, map[string]string{
				"az":             string(k + 65),
				"cidrblockInner": v[t].String(),
				"function":       t,
			})
		}
	}

	return buf.String(), nil
}