package ec2

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// FindCarrierGatewayByID returns the carrier gateway corresponding to the specified identifier.
// Returns nil and potentially an error if no carrier gateway is found.
func FindCarrierGatewayByID(conn *ec2.EC2, id string) (*ec2.CarrierGateway, error) {
	input := &ec2.DescribeCarrierGatewaysInput{
		CarrierGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeCarrierGateways(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CarrierGateways) == 0 {
		return nil, nil
	}

	return output.CarrierGateways[0], nil
}

func FindClientVPNEndpoint(conn *ec2.EC2, input *ec2.DescribeClientVpnEndpointsInput) (*ec2.ClientVpnEndpoint, error) {
	output, err := FindClientVPNEndpoints(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindClientVPNEndpoints(conn *ec2.EC2, input *ec2.DescribeClientVpnEndpointsInput) ([]*ec2.ClientVpnEndpoint, error) {
	var output []*ec2.ClientVpnEndpoint

	err := conn.DescribeClientVpnEndpointsPages(input, func(page *ec2.DescribeClientVpnEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ClientVpnEndpoints {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidClientVpnEndpointIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindClientVPNEndpointByID(conn *ec2.EC2, id string) (*ec2.ClientVpnEndpoint, error) {
	input := &ec2.DescribeClientVpnEndpointsInput{
		ClientVpnEndpointIds: aws.StringSlice([]string{id}),
	}

	output, err := FindClientVPNEndpoint(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.Status.Code); state == ec2.ClientVpnEndpointStatusCodeDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.ClientVpnEndpointId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindClientVPNEndpointClientConnectResponseOptionsByID(conn *ec2.EC2, id string) (*ec2.ClientConnectResponseOptions, error) {
	output, err := FindClientVPNEndpointByID(conn, id)

	if err != nil {
		return nil, err
	}

	if output.ClientConnectOptions == nil || output.ClientConnectOptions.Status == nil {
		return nil, tfresource.NewEmptyResultError(id)
	}

	return output.ClientConnectOptions, nil
}

func FindClientVPNAuthorizationRule(conn *ec2.EC2, input *ec2.DescribeClientVpnAuthorizationRulesInput) (*ec2.AuthorizationRule, error) {
	output, err := FindClientVPNAuthorizationRules(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindClientVPNAuthorizationRules(conn *ec2.EC2, input *ec2.DescribeClientVpnAuthorizationRulesInput) ([]*ec2.AuthorizationRule, error) {
	var output []*ec2.AuthorizationRule

	err := conn.DescribeClientVpnAuthorizationRulesPages(input, func(page *ec2.DescribeClientVpnAuthorizationRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AuthorizationRules {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidClientVpnEndpointIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindClientVPNAuthorizationRuleByThreePartKey(conn *ec2.EC2, endpointID, targetNetworkCIDR, accessGroupID string) (*ec2.AuthorizationRule, error) {
	filters := map[string]string{
		"destination-cidr": targetNetworkCIDR,
	}
	if accessGroupID != "" {
		filters["group-id"] = accessGroupID
	}
	input := &ec2.DescribeClientVpnAuthorizationRulesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters:             BuildAttributeFilterList(filters),
	}

	return FindClientVPNAuthorizationRule(conn, input)
}

func FindClientVPNNetworkAssociation(conn *ec2.EC2, input *ec2.DescribeClientVpnTargetNetworksInput) (*ec2.TargetNetwork, error) {
	output, err := FindClientVPNNetworkAssociations(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindClientVPNNetworkAssociations(conn *ec2.EC2, input *ec2.DescribeClientVpnTargetNetworksInput) ([]*ec2.TargetNetwork, error) {
	var output []*ec2.TargetNetwork

	err := conn.DescribeClientVpnTargetNetworksPages(input, func(page *ec2.DescribeClientVpnTargetNetworksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ClientVpnTargetNetworks {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidClientVpnEndpointIdNotFound, ErrCodeInvalidClientVpnAssociationIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindClientVPNNetworkAssociationByIDs(conn *ec2.EC2, associationID, endpointID string) (*ec2.TargetNetwork, error) {
	input := &ec2.DescribeClientVpnTargetNetworksInput{
		AssociationIds:      aws.StringSlice([]string{associationID}),
		ClientVpnEndpointId: aws.String(endpointID),
	}

	output, err := FindClientVPNNetworkAssociation(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.Status.Code); state == ec2.AssociationStatusCodeDisassociated {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.ClientVpnEndpointId) != endpointID || aws.StringValue(output.AssociationId) != associationID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindClientVPNRoute(conn *ec2.EC2, input *ec2.DescribeClientVpnRoutesInput) (*ec2.ClientVpnRoute, error) {
	output, err := FindClientVPNRoutes(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindClientVPNRoutes(conn *ec2.EC2, input *ec2.DescribeClientVpnRoutesInput) ([]*ec2.ClientVpnRoute, error) {
	var output []*ec2.ClientVpnRoute

	err := conn.DescribeClientVpnRoutesPages(input, func(page *ec2.DescribeClientVpnRoutesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Routes {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidClientVpnEndpointIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindClientVPNRouteByThreePartKey(conn *ec2.EC2, endpointID, targetSubnetID, destinationCIDR string) (*ec2.ClientVpnRoute, error) {
	input := &ec2.DescribeClientVpnRoutesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters: BuildAttributeFilterList(map[string]string{
			"destination-cidr": destinationCIDR,
			"target-subnet":    targetSubnetID,
		}),
	}

	return FindClientVPNRoute(conn, input)
}

func FindCOIPPools(conn *ec2.EC2, input *ec2.DescribeCoipPoolsInput) ([]*ec2.CoipPool, error) {
	var output []*ec2.CoipPool

	err := conn.DescribeCoipPoolsPages(input, func(page *ec2.DescribeCoipPoolsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CoipPools {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidPoolIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindCOIPPool(conn *ec2.EC2, input *ec2.DescribeCoipPoolsInput) (*ec2.CoipPool, error) {
	output, err := FindCOIPPools(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindEBSVolumes(conn *ec2.EC2, input *ec2.DescribeVolumesInput) ([]*ec2.Volume, error) {
	var output []*ec2.Volume

	err := conn.DescribeVolumesPages(input, func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Volumes {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVolumeNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindEBSVolume(conn *ec2.EC2, input *ec2.DescribeVolumesInput) (*ec2.Volume, error) {
	output, err := FindEBSVolumes(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindEIPs(conn *ec2.EC2, input *ec2.DescribeAddressesInput) ([]*ec2.Address, error) {
	var addresses []*ec2.Address

	output, err := conn.DescribeAddresses(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidAddressNotFound, ErrCodeInvalidAllocationIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	for _, v := range output.Addresses {
		if v != nil {
			addresses = append(addresses, v)
		}
	}

	return addresses, nil
}

func FindHostByID(conn *ec2.EC2, id string) (*ec2.Host, error) {
	input := &ec2.DescribeHostsInput{
		HostIds: aws.StringSlice([]string{id}),
	}

	return FindHost(conn, input)
}

func FindHostByIDAndFilters(conn *ec2.EC2, id string, filters []*ec2.Filter) (*ec2.Host, error) {
	input := &ec2.DescribeHostsInput{}

	if id != "" {
		input.HostIds = aws.StringSlice([]string{id})
	}

	if len(filters) > 0 {
		input.Filter = filters
	}

	return FindHost(conn, input)
}

func FindHost(conn *ec2.EC2, input *ec2.DescribeHostsInput) (*ec2.Host, error) {
	output, err := conn.DescribeHosts(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidHostIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Hosts) == 0 || output.Hosts[0] == nil || output.Hosts[0].HostProperties == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Hosts); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	host := output.Hosts[0]

	if state := aws.StringValue(host.State); state == ec2.AllocationStateReleased || state == ec2.AllocationStateReleasedPermanentFailure {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return host, nil
}

func FindInstances(conn *ec2.EC2, input *ec2.DescribeInstancesInput) ([]*ec2.Instance, error) {
	var output []*ec2.Instance

	err := conn.DescribeInstancesPages(input, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Reservations {
			if v != nil {
				for _, v := range v.Instances {
					if v != nil {
						output = append(output, v)
					}
				}
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidInstanceIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindInstance(conn *ec2.EC2, input *ec2.DescribeInstancesInput) (*ec2.Instance, error) {
	output, err := FindInstances(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindInstanceByID(conn *ec2.EC2, id string) (*ec2.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindInstance(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State.Name); state == ec2.InstanceStateNameTerminated {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.InstanceId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindLocalGatewayRouteTables(conn *ec2.EC2, input *ec2.DescribeLocalGatewayRouteTablesInput) ([]*ec2.LocalGatewayRouteTable, error) {
	var output []*ec2.LocalGatewayRouteTable

	err := conn.DescribeLocalGatewayRouteTablesPages(input, func(page *ec2.DescribeLocalGatewayRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LocalGatewayRouteTables {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindLocalGatewayRouteTable(conn *ec2.EC2, input *ec2.DescribeLocalGatewayRouteTablesInput) (*ec2.LocalGatewayRouteTable, error) {
	output, err := FindLocalGatewayRouteTables(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindLocalGatewayVirtualInterfaceGroups(conn *ec2.EC2, input *ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput) ([]*ec2.LocalGatewayVirtualInterfaceGroup, error) {
	var output []*ec2.LocalGatewayVirtualInterfaceGroup

	err := conn.DescribeLocalGatewayVirtualInterfaceGroupsPages(input, func(page *ec2.DescribeLocalGatewayVirtualInterfaceGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LocalGatewayVirtualInterfaceGroups {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindLocalGatewayVirtualInterfaceGroup(conn *ec2.EC2, input *ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput) (*ec2.LocalGatewayVirtualInterfaceGroup, error) {
	output, err := FindLocalGatewayVirtualInterfaceGroups(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindLocalGateways(conn *ec2.EC2, input *ec2.DescribeLocalGatewaysInput) ([]*ec2.LocalGateway, error) {
	var output []*ec2.LocalGateway

	err := conn.DescribeLocalGatewaysPages(input, func(page *ec2.DescribeLocalGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LocalGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindLocalGateway(conn *ec2.EC2, input *ec2.DescribeLocalGatewaysInput) (*ec2.LocalGateway, error) {
	output, err := FindLocalGateways(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindNetworkACL(conn *ec2.EC2, input *ec2.DescribeNetworkAclsInput) (*ec2.NetworkAcl, error) {
	output, err := FindNetworkACLs(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindNetworkACLs(conn *ec2.EC2, input *ec2.DescribeNetworkAclsInput) ([]*ec2.NetworkAcl, error) {
	var output []*ec2.NetworkAcl

	err := conn.DescribeNetworkAclsPages(input, func(page *ec2.DescribeNetworkAclsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkAcls {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidNetworkAclIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNetworkACLByID(conn *ec2.EC2, id string) (*ec2.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		NetworkAclIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkACL(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.NetworkAclId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkACLAssociationByID(conn *ec2.EC2, associationID string) (*ec2.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.association-id": associationID,
		}),
	}

	output, err := FindNetworkACL(conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Associations {
		if aws.StringValue(v.NetworkAclAssociationId) == associationID {
			return v, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindNetworkACLAssociationBySubnetID(conn *ec2.EC2, subnetID string) (*ec2.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.subnet-id": subnetID,
		}),
	}

	output, err := FindNetworkACL(conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Associations {
		if aws.StringValue(v.SubnetId) == subnetID {
			return v, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindNetworkACLEntryByThreePartKey(conn *ec2.EC2, naclID string, egress bool, ruleNumber int) (*ec2.NetworkAclEntry, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"entry.egress":      strconv.FormatBool(egress),
			"entry.rule-number": strconv.Itoa(ruleNumber),
		}),
		NetworkAclIds: aws.StringSlice([]string{naclID}),
	}

	output, err := FindNetworkACL(conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Entries {
		if aws.BoolValue(v.Egress) == egress && aws.Int64Value(v.RuleNumber) == int64(ruleNumber) {
			return v, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindNetworkInterface(conn *ec2.EC2, input *ec2.DescribeNetworkInterfacesInput) (*ec2.NetworkInterface, error) {
	output, err := FindNetworkInterfaces(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindNetworkInterfaces(conn *ec2.EC2, input *ec2.DescribeNetworkInterfacesInput) ([]*ec2.NetworkInterface, error) {
	var output []*ec2.NetworkInterface

	err := conn.DescribeNetworkInterfacesPages(input, func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkInterfaces {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidNetworkInterfaceIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNetworkInterfaceByID(conn *ec2.EC2, id string) (*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkInterface(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.NetworkInterfaceId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkInterfacesByAttachmentInstanceOwnerIDAndDescription(conn *ec2.EC2, attachmentInstanceOwnerID, description string) ([]*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"attachment.instance-owner-id": attachmentInstanceOwnerID,
			"description":                  description,
		}),
	}

	return FindNetworkInterfaces(conn, input)
}

func FindNetworkInterfaceAttachmentByID(conn *ec2.EC2, id string) (*ec2.NetworkInterfaceAttachment, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"attachment.attachment-id": id,
		}),
	}

	networkInterface, err := FindNetworkInterface(conn, input)

	if err != nil {
		return nil, err
	}

	if networkInterface.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return networkInterface.Attachment, nil
}

func FindNetworkInterfaceSecurityGroup(conn *ec2.EC2, networkInterfaceID string, securityGroupID string) (*ec2.GroupIdentifier, error) {
	networkInterface, err := FindNetworkInterfaceByID(conn, networkInterfaceID)

	if err != nil {
		return nil, err
	}

	for _, groupIdentifier := range networkInterface.Groups {
		if aws.StringValue(groupIdentifier.GroupId) == securityGroupID {
			return groupIdentifier, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Network Interface (%s) Security Group (%s) not found", networkInterfaceID, securityGroupID),
	}
}

// FindMainRouteTableAssociationByID returns the main route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func FindMainRouteTableAssociationByID(conn *ec2.EC2, associationID string) (*ec2.RouteTableAssociation, error) {
	association, err := FindRouteTableAssociationByID(conn, associationID)

	if err != nil {
		return nil, err
	}

	if !aws.BoolValue(association.Main) {
		return nil, &resource.NotFoundError{
			Message: fmt.Sprintf("%s is not the association with the main route table", associationID),
		}
	}

	return association, err
}

// FindMainRouteTableAssociationByVPCID returns the main route table association for the specified VPC.
// Returns NotFoundError if no route table association is found.
func FindMainRouteTableAssociationByVPCID(conn *ec2.EC2, vpcID string) (*ec2.RouteTableAssociation, error) {
	routeTable, err := FindMainRouteTableByVPCID(conn, vpcID)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.BoolValue(association.Main) {
			if association.AssociationState != nil {
				if state := aws.StringValue(association.AssociationState.State); state == ec2.RouteTableAssociationStateCodeDisassociated {
					continue
				}
			}

			return association, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

// FindRouteTableAssociationByID returns the route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func FindRouteTableAssociationByID(conn *ec2.EC2, associationID string) (*ec2.RouteTableAssociation, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.route-table-association-id": associationID,
		}),
	}

	routeTable, err := FindRouteTable(conn, input)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.StringValue(association.RouteTableAssociationId) == associationID {
			if association.AssociationState != nil {
				if state := aws.StringValue(association.AssociationState.State); state == ec2.RouteTableAssociationStateCodeDisassociated {
					return nil, &resource.NotFoundError{Message: state}
				}
			}

			return association, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

// FindMainRouteTableByVPCID returns the main route table for the specified VPC.
// Returns NotFoundError if no route table is found.
func FindMainRouteTableByVPCID(conn *ec2.EC2, vpcID string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.main": "true",
			"vpc-id":           vpcID,
		}),
	}

	return FindRouteTable(conn, input)
}

// FindRouteTableByID returns the route table corresponding to the specified identifier.
// Returns NotFoundError if no route table is found.
func FindRouteTableByID(conn *ec2.EC2, routeTableID string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		RouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	return FindRouteTable(conn, input)
}

func FindRouteTable(conn *ec2.EC2, input *ec2.DescribeRouteTablesInput) (*ec2.RouteTable, error) {
	output, err := FindRouteTables(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindRouteTables(conn *ec2.EC2, input *ec2.DescribeRouteTablesInput) ([]*ec2.RouteTable, error) {
	var output []*ec2.RouteTable

	err := conn.DescribeRouteTablesPages(input, func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, table := range page.RouteTables {
			if table == nil {
				continue
			}

			output = append(output, table)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// RouteFinder returns the route corresponding to the specified destination.
// Returns NotFoundError if no route is found.
type RouteFinder func(*ec2.EC2, string, string) (*ec2.Route, error)

// FindRouteByIPv4Destination returns the route corresponding to the specified IPv4 destination.
// Returns NotFoundError if no route is found.
func FindRouteByIPv4Destination(conn *ec2.EC2, routeTableID, destinationCidr string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if verify.CIDRBlocksEqual(aws.StringValue(route.DestinationCidrBlock), destinationCidr) {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with IPv4 destination (%s) not found", routeTableID, destinationCidr),
	}
}

// FindRouteByIPv6Destination returns the route corresponding to the specified IPv6 destination.
// Returns NotFoundError if no route is found.
func FindRouteByIPv6Destination(conn *ec2.EC2, routeTableID, destinationIpv6Cidr string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if verify.CIDRBlocksEqual(aws.StringValue(route.DestinationIpv6CidrBlock), destinationIpv6Cidr) {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with IPv6 destination (%s) not found", routeTableID, destinationIpv6Cidr),
	}
}

// FindRouteByPrefixListIDDestination returns the route corresponding to the specified prefix list destination.
// Returns NotFoundError if no route is found.
func FindRouteByPrefixListIDDestination(conn *ec2.EC2, routeTableID, prefixListID string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(conn, routeTableID)
	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if aws.StringValue(route.DestinationPrefixListId) == prefixListID {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with Prefix List ID destination (%s) not found", routeTableID, prefixListID),
	}
}

func FindSecurityGroupByID(conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSecurityGroup(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.GroupId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

// FindSecurityGroupByNameAndVPCID looks up a security group by name and VPC ID. Returns a resource.NotFoundError if not found.
func FindSecurityGroupByNameAndVPCID(conn *ec2.EC2, name, vpcID string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildAttributeFilterList(
			map[string]string{
				"group-name": name,
				"vpc-id":     vpcID,
			},
		),
	}
	return FindSecurityGroup(conn, input)
}

// FindSecurityGroup looks up a security group using an ec2.DescribeSecurityGroupsInput. Returns a resource.NotFoundError if not found.
func FindSecurityGroup(conn *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) (*ec2.SecurityGroup, error) {
	output, err := FindSecurityGroups(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindSecurityGroups(conn *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) ([]*ec2.SecurityGroup, error) {
	var output []*ec2.SecurityGroup

	err := conn.DescribeSecurityGroupsPages(input, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SecurityGroups {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidGroupNotFound, ErrCodeInvalidSecurityGroupIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindSpotInstanceRequestByID looks up a SpotInstanceRequest by ID. When not found, returns nil and potentially an API error.
func FindSpotInstanceRequestByID(conn *ec2.EC2, id string) (*ec2.SpotInstanceRequest, error) {
	input := &ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeSpotInstanceRequests(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, spotInstanceRequest := range output.SpotInstanceRequests {
		if spotInstanceRequest == nil {
			continue
		}

		if aws.StringValue(spotInstanceRequest.SpotInstanceRequestId) != id {
			continue
		}

		return spotInstanceRequest, nil
	}

	return nil, nil
}

func FindSubnetByID(conn *ec2.EC2, id string) (*ec2.Subnet, error) {
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSubnet(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.SubnetId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSubnet(conn *ec2.EC2, input *ec2.DescribeSubnetsInput) (*ec2.Subnet, error) {
	output, err := FindSubnets(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindSubnets(conn *ec2.EC2, input *ec2.DescribeSubnetsInput) ([]*ec2.Subnet, error) {
	var output []*ec2.Subnet

	err := conn.DescribeSubnetsPages(input, func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Subnets {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSubnetIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSubnetCidrReservationBySubnetIDAndReservationID(conn *ec2.EC2, subnetID, reservationID string) (*ec2.SubnetCidrReservation, error) {
	input := &ec2.GetSubnetCidrReservationsInput{
		SubnetId: aws.String(subnetID),
	}

	output, err := conn.GetSubnetCidrReservations(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSubnetIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || (len(output.SubnetIpv4CidrReservations) == 0 && len(output.SubnetIpv6CidrReservations) == 0) {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, r := range output.SubnetIpv4CidrReservations {
		if aws.StringValue(r.SubnetCidrReservationId) == reservationID {
			return r, nil
		}
	}
	for _, r := range output.SubnetIpv6CidrReservations {
		if aws.StringValue(r.SubnetCidrReservationId) == reservationID {
			return r, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError:   err,
		LastRequest: input,
	}
}

func FindSubnetIPv6CIDRBlockAssociationByID(conn *ec2.EC2, associationID string) (*ec2.SubnetIpv6CidrBlockAssociation, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"ipv6-cidr-block-association.association-id": associationID,
		}),
	}

	output, err := FindSubnet(conn, input)

	if err != nil {
		return nil, err
	}

	for _, association := range output.Ipv6CidrBlockAssociationSet {
		if aws.StringValue(association.AssociationId) == associationID {
			if state := aws.StringValue(association.Ipv6CidrBlockState.State); state == ec2.SubnetCidrBlockStateCodeDisassociated {
				return nil, &resource.NotFoundError{Message: state}
			}

			return association, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindVPCAttribute(conn *ec2.EC2, vpcID string, attribute string) (bool, error) {
	input := &ec2.DescribeVpcAttributeInput{
		Attribute: aws.String(attribute),
		VpcId:     aws.String(vpcID),
	}

	output, err := conn.DescribeVpcAttribute(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpcIDNotFound) {
		return false, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return false, err
	}

	if output == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	var v *ec2.AttributeBooleanValue
	switch attribute {
	case ec2.VpcAttributeNameEnableDnsHostnames:
		v = output.EnableDnsHostnames
	case ec2.VpcAttributeNameEnableDnsSupport:
		v = output.EnableDnsSupport
	default:
		return false, fmt.Errorf("unsupported VPC attribute: %s", attribute)
	}

	if v == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	return aws.BoolValue(v.Value), nil
}

func FindVPCClassicLinkEnabled(conn *ec2.EC2, vpcID string) (bool, error) {
	input := &ec2.DescribeVpcClassicLinkInput{
		VpcIds: aws.StringSlice([]string{vpcID}),
	}

	output, err := conn.DescribeVpcClassicLink(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpcIDNotFound, ErrCodeUnsupportedOperation) {
		return false, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return false, err
	}

	if output == nil || len(output.Vpcs) == 0 || output.Vpcs[0] == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Vpcs); count > 1 {
		return false, tfresource.NewTooManyResultsError(count, input)
	}

	vpc := output.Vpcs[0]

	// Eventual consistency check.
	if aws.StringValue(vpc.VpcId) != vpcID {
		return false, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return aws.BoolValue(vpc.ClassicLinkEnabled), nil
}

func FindVPCClassicLinkDnsSupported(conn *ec2.EC2, vpcID string) (bool, error) {
	input := &ec2.DescribeVpcClassicLinkDnsSupportInput{
		VpcIds: aws.StringSlice([]string{vpcID}),
	}

	output, err := conn.DescribeVpcClassicLinkDnsSupport(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpcIDNotFound, ErrCodeUnsupportedOperation) ||
		tfawserr.ErrMessageContains(err, ErrCodeAuthFailure, "This request has been administratively disabled") {
		return false, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return false, err
	}

	if output == nil || len(output.Vpcs) == 0 || output.Vpcs[0] == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Vpcs); count > 1 {
		return false, tfresource.NewTooManyResultsError(count, input)
	}

	vpc := output.Vpcs[0]

	// Eventual consistency check.
	if aws.StringValue(vpc.VpcId) != vpcID {
		return false, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return aws.BoolValue(vpc.ClassicLinkDnsSupported), nil
}

func FindVPC(conn *ec2.EC2, input *ec2.DescribeVpcsInput) (*ec2.Vpc, error) {
	output, err := FindVPCs(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindVPCs(conn *ec2.EC2, input *ec2.DescribeVpcsInput) ([]*ec2.Vpc, error) {
	var output []*ec2.Vpc

	err := conn.DescribeVpcsPages(input, func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Vpcs {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpcIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCByID(conn *ec2.EC2, id string) (*ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPC(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpcId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCDHCPOptionsAssociation(conn *ec2.EC2, vpcID string, dhcpOptionsID string) error {
	vpc, err := FindVPCByID(conn, vpcID)

	if err != nil {
		return err
	}

	if aws.StringValue(vpc.DhcpOptionsId) != dhcpOptionsID {
		return &resource.NotFoundError{
			LastError: fmt.Errorf("EC2 VPC (%s) DHCP Options Set (%s) Association not found", vpcID, dhcpOptionsID),
		}
	}

	return nil
}

func FindVPCCIDRBlockAssociationByID(conn *ec2.EC2, id string) (*ec2.VpcCidrBlockAssociation, *ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"cidr-block-association.association-id": id,
		}),
	}

	vpc, err := FindVPC(conn, input)

	if err != nil {
		return nil, nil, err
	}

	for _, association := range vpc.CidrBlockAssociationSet {
		if aws.StringValue(association.AssociationId) == id {
			if state := aws.StringValue(association.CidrBlockState.State); state == ec2.VpcCidrBlockStateCodeDisassociated {
				return nil, nil, &resource.NotFoundError{Message: state}
			}

			return association, vpc, nil
		}
	}

	return nil, nil, &resource.NotFoundError{}
}

func FindVPCIPv6CIDRBlockAssociationByID(conn *ec2.EC2, id string) (*ec2.VpcIpv6CidrBlockAssociation, *ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"ipv6-cidr-block-association.association-id": id,
		}),
	}

	vpc, err := FindVPC(conn, input)

	if err != nil {
		return nil, nil, err
	}

	for _, association := range vpc.Ipv6CidrBlockAssociationSet {
		if aws.StringValue(association.AssociationId) == id {
			if state := aws.StringValue(association.Ipv6CidrBlockState.State); state == ec2.VpcCidrBlockStateCodeDisassociated {
				return nil, nil, &resource.NotFoundError{Message: state}
			}

			return association, vpc, nil
		}
	}

	return nil, nil, &resource.NotFoundError{}
}

func FindVPCDefaultNetworkACL(conn *ec2.EC2, id string) (*ec2.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"default": "true",
			"vpc-id":  id,
		}),
	}

	return FindNetworkACL(conn, input)
}

func FindVPCDefaultSecurityGroup(conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"group-name": DefaultSecurityGroupName,
			"vpc-id":     id,
		}),
	}

	return FindSecurityGroup(conn, input)
}

func FindVPCMainRouteTable(conn *ec2.EC2, id string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.main": "true",
			"vpc-id":           id,
		}),
	}

	return FindRouteTable(conn, input)
}

// FindVPCEndpointByID returns the VPC endpoint corresponding to the specified identifier.
// Returns NotFoundError if no VPC endpoint is found.
func FindVPCEndpointByID(conn *ec2.EC2, vpcEndpointID string) (*ec2.VpcEndpoint, error) {
	input := &ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: aws.StringSlice([]string{vpcEndpointID}),
	}

	output, err := FindVPCEndpoint(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == VpcEndpointStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpcEndpointId) != vpcEndpointID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCEndpoint(conn *ec2.EC2, input *ec2.DescribeVpcEndpointsInput) (*ec2.VpcEndpoint, error) {
	output, err := conn.DescribeVpcEndpoints(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpcEndpointIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpcEndpoints) == 0 || output.VpcEndpoints[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VpcEndpoints[0], nil
}

// FindVPCEndpointRouteTableAssociationExists returns NotFoundError if no association for the specified VPC endpoint and route table IDs is found.
func FindVPCEndpointRouteTableAssociationExists(conn *ec2.EC2, vpcEndpointID string, routeTableID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointRouteTableID := range vpcEndpoint.RouteTableIds {
		if aws.StringValue(vpcEndpointRouteTableID) == routeTableID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint Route Table Association (%s/%s) not found", vpcEndpointID, routeTableID),
	}
}

// FindVPCEndpointSubnetAssociationExists returns NotFoundError if no association for the specified VPC endpoint and subnet IDs is found.
func FindVPCEndpointSubnetAssociationExists(conn *ec2.EC2, vpcEndpointID string, subnetID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointSubnetID := range vpcEndpoint.SubnetIds {
		if aws.StringValue(vpcEndpointSubnetID) == subnetID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Subnet (%s) Association not found", vpcEndpointID, subnetID),
	}
}

func FindVPCPeeringConnection(conn *ec2.EC2, input *ec2.DescribeVpcPeeringConnectionsInput) (*ec2.VpcPeeringConnection, error) {
	output, err := FindVPCPeeringConnections(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindVPCPeeringConnections(conn *ec2.EC2, input *ec2.DescribeVpcPeeringConnectionsInput) ([]*ec2.VpcPeeringConnection, error) {
	var output []*ec2.VpcPeeringConnection

	err := conn.DescribeVpcPeeringConnectionsPages(input, func(page *ec2.DescribeVpcPeeringConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VpcPeeringConnections {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpcPeeringConnectionIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCPeeringConnectionByID(conn *ec2.EC2, id string) (*ec2.VpcPeeringConnection, error) {
	input := &ec2.DescribeVpcPeeringConnectionsInput{
		VpcPeeringConnectionIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPCPeeringConnection(conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/vpc/latest/peering/vpc-peering-basics.html#vpc-peering-lifecycle.
	switch statusCode := aws.StringValue(output.Status.Code); statusCode {
	case ec2.VpcPeeringConnectionStateReasonCodeDeleted,
		ec2.VpcPeeringConnectionStateReasonCodeExpired,
		ec2.VpcPeeringConnectionStateReasonCodeFailed,
		ec2.VpcPeeringConnectionStateReasonCodeRejected:
		return nil, &resource.NotFoundError{
			Message:     statusCode,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpcPeeringConnectionId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

// FindVPNGatewayRoutePropagationExists returns NotFoundError if no route propagation for the specified VPN gateway is found.
func FindVPNGatewayRoutePropagationExists(conn *ec2.EC2, routeTableID, gatewayID string) error {
	routeTable, err := FindRouteTableByID(conn, routeTableID)

	if err != nil {
		return err
	}

	for _, v := range routeTable.PropagatingVgws {
		if aws.StringValue(v.GatewayId) == gatewayID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("Route Table (%s) VPN Gateway (%s) route propagation not found", routeTableID, gatewayID),
	}
}

func FindVPNGatewayVPCAttachment(conn *ec2.EC2, vpnGatewayID, vpcID string) (*ec2.VpcAttachment, error) {
	vpnGateway, err := FindVPNGatewayByID(conn, vpnGatewayID)

	if err != nil {
		return nil, err
	}

	for _, vpcAttachment := range vpnGateway.VpcAttachments {
		if aws.StringValue(vpcAttachment.VpcId) == vpcID {
			if state := aws.StringValue(vpcAttachment.State); state == ec2.AttachmentStatusDetached {
				return nil, &resource.NotFoundError{
					Message:     state,
					LastRequest: vpcID,
				}
			}

			return vpcAttachment, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(vpcID)
}

func FindVPNGatewayByID(conn *ec2.EC2, id string) (*ec2.VpnGateway, error) {
	input := &ec2.DescribeVpnGatewaysInput{
		VpnGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPNGateway(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.VpnStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpnGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPNGateway(conn *ec2.EC2, input *ec2.DescribeVpnGatewaysInput) (*ec2.VpnGateway, error) {
	output, err := conn.DescribeVpnGateways(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpnGatewayIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpnGateways) == 0 || output.VpnGateways[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.VpnGateways); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.VpnGateways[0], nil
}

func FindCustomerGatewayByID(conn *ec2.EC2, id string) (*ec2.CustomerGateway, error) {
	input := &ec2.DescribeCustomerGatewaysInput{
		CustomerGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindCustomerGateway(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == CustomerGatewayStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.CustomerGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindCustomerGateway(conn *ec2.EC2, input *ec2.DescribeCustomerGatewaysInput) (*ec2.CustomerGateway, error) {
	output, err := conn.DescribeCustomerGateways(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidCustomerGatewayIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CustomerGateways) == 0 || output.CustomerGateways[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.CustomerGateways); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.CustomerGateways[0], nil
}

func FindVPNConnectionByID(conn *ec2.EC2, id string) (*ec2.VpnConnection, error) {
	input := &ec2.DescribeVpnConnectionsInput{
		VpnConnectionIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPNConnection(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.VpnStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpnConnectionId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPNConnection(conn *ec2.EC2, input *ec2.DescribeVpnConnectionsInput) (*ec2.VpnConnection, error) {
	output, err := conn.DescribeVpnConnections(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpnConnectionIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpnConnections) == 0 || output.VpnConnections[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.VpnConnections); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.VpnConnections[0], nil
}

func FindVPNConnectionRouteByVPNConnectionIDAndCIDR(conn *ec2.EC2, vpnConnectionID, cidrBlock string) (*ec2.VpnStaticRoute, error) {
	input := &ec2.DescribeVpnConnectionsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"route.destination-cidr-block": cidrBlock,
			"vpn-connection-id":            vpnConnectionID,
		}),
	}

	output, err := FindVPNConnection(conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Routes {
		if aws.StringValue(v.DestinationCidrBlock) == cidrBlock && aws.StringValue(v.State) != ec2.VpnStateDeleted {
			return v, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("EC2 VPN Connection (%s) Route (%s) not found", vpnConnectionID, cidrBlock),
	}
}

func FindTransitGateway(conn *ec2.EC2, input *ec2.DescribeTransitGatewaysInput) (*ec2.TransitGateway, error) {
	output, err := FindTransitGateways(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Options == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGateways(conn *ec2.EC2, input *ec2.DescribeTransitGatewaysInput) ([]*ec2.TransitGateway, error) {
	var output []*ec2.TransitGateway

	err := conn.DescribeTransitGatewaysPages(input, func(page *ec2.DescribeTransitGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidTransitGatewayIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayByID(conn *ec2.EC2, id string) (*ec2.TransitGateway, error) {
	input := &ec2.DescribeTransitGatewaysInput{
		TransitGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGateway(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayAttachment(conn *ec2.EC2, input *ec2.DescribeTransitGatewayAttachmentsInput) (*ec2.TransitGatewayAttachment, error) {
	output, err := FindTransitGatewayAttachments(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayAttachments(conn *ec2.EC2, input *ec2.DescribeTransitGatewayAttachmentsInput) ([]*ec2.TransitGatewayAttachment, error) {
	var output []*ec2.TransitGatewayAttachment

	err := conn.DescribeTransitGatewayAttachmentsPages(input, func(page *ec2.DescribeTransitGatewayAttachmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayAttachments {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayAttachmentByID(conn *ec2.EC2, id string) (*ec2.TransitGatewayAttachment, error) {
	input := &ec2.DescribeTransitGatewayAttachmentsInput{
		TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayAttachment(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayConnect(conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectsInput) (*ec2.TransitGatewayConnect, error) {
	output, err := FindTransitGatewayConnects(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Options == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayConnects(conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectsInput) ([]*ec2.TransitGatewayConnect, error) {
	var output []*ec2.TransitGatewayConnect

	err := conn.DescribeTransitGatewayConnectsPages(input, func(page *ec2.DescribeTransitGatewayConnectsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayConnects {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayConnectByID(conn *ec2.EC2, id string) (*ec2.TransitGatewayConnect, error) {
	input := &ec2.DescribeTransitGatewayConnectsInput{
		TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayConnect(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayAttachmentStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayConnectPeer(conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectPeersInput) (*ec2.TransitGatewayConnectPeer, error) {
	output, err := FindTransitGatewayConnectPeers(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].ConnectPeerConfiguration == nil ||
		len(output[0].ConnectPeerConfiguration.BgpConfigurations) == 0 || output[0].ConnectPeerConfiguration.BgpConfigurations[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayConnectPeers(conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectPeersInput) ([]*ec2.TransitGatewayConnectPeer, error) {
	var output []*ec2.TransitGatewayConnectPeer

	err := conn.DescribeTransitGatewayConnectPeersPages(input, func(page *ec2.DescribeTransitGatewayConnectPeersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayConnectPeers {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidTransitGatewayConnectPeerIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayConnectPeerByID(conn *ec2.EC2, id string) (*ec2.TransitGatewayConnectPeer, error) {
	input := &ec2.DescribeTransitGatewayConnectPeersInput{
		TransitGatewayConnectPeerIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayConnectPeer(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayConnectPeerStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayConnectPeerId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayMulticastDomain(conn *ec2.EC2, input *ec2.DescribeTransitGatewayMulticastDomainsInput) (*ec2.TransitGatewayMulticastDomain, error) {
	output, err := FindTransitGatewayMulticastDomains(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Options == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayMulticastDomains(conn *ec2.EC2, input *ec2.DescribeTransitGatewayMulticastDomainsInput) ([]*ec2.TransitGatewayMulticastDomain, error) {
	var output []*ec2.TransitGatewayMulticastDomain

	err := conn.DescribeTransitGatewayMulticastDomainsPages(input, func(page *ec2.DescribeTransitGatewayMulticastDomainsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayMulticastDomains {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayMulticastDomainByID(conn *ec2.EC2, id string) (*ec2.TransitGatewayMulticastDomain, error) {
	input := &ec2.DescribeTransitGatewayMulticastDomainsInput{
		TransitGatewayMulticastDomainIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayMulticastDomain(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayMulticastDomainStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayMulticastDomainId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayMulticastDomainAssociation(conn *ec2.EC2, input *ec2.GetTransitGatewayMulticastDomainAssociationsInput) (*ec2.TransitGatewayMulticastDomainAssociation, error) {
	output, err := FindTransitGatewayMulticastDomainAssociations(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Subnet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayMulticastDomainAssociations(conn *ec2.EC2, input *ec2.GetTransitGatewayMulticastDomainAssociationsInput) ([]*ec2.TransitGatewayMulticastDomainAssociation, error) {
	var output []*ec2.TransitGatewayMulticastDomainAssociation

	err := conn.GetTransitGatewayMulticastDomainAssociationsPages(input, func(page *ec2.GetTransitGatewayMulticastDomainAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.MulticastDomainAssociations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayMulticastDomainAssociationByThreePartKey(conn *ec2.EC2, multicastDomainID, attachmentID, subnetID string) (*ec2.TransitGatewayMulticastDomainAssociation, error) {
	input := &ec2.GetTransitGatewayMulticastDomainAssociationsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"subnet-id":                     subnetID,
			"transit-gateway-attachment-id": attachmentID,
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := FindTransitGatewayMulticastDomainAssociation(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.Subnet.State); state == ec2.TransitGatewayMulitcastDomainAssociationStateDisassociated {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != attachmentID || aws.StringValue(output.Subnet.SubnetId) != subnetID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayMulticastGroups(conn *ec2.EC2, input *ec2.SearchTransitGatewayMulticastGroupsInput) ([]*ec2.TransitGatewayMulticastGroup, error) {
	var output []*ec2.TransitGatewayMulticastGroup

	err := conn.SearchTransitGatewayMulticastGroupsPages(input, func(page *ec2.SearchTransitGatewayMulticastGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.MulticastGroups {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayMulticastGroupMemberByThreePartKey(conn *ec2.EC2, multicastDomainID, groupIPAddress, eniID string) (*ec2.TransitGatewayMulticastGroup, error) {
	input := &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"group-ip-address": groupIPAddress,
			"is-group-member":  "true",
			"is-group-source":  "false",
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := FindTransitGatewayMulticastGroups(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output {
		if aws.StringValue(v.NetworkInterfaceId) == eniID {
			// Eventual consistency check.
			if aws.StringValue(v.GroupIpAddress) != groupIPAddress || !aws.BoolValue(v.GroupMember) {
				return nil, &resource.NotFoundError{
					LastRequest: input,
				}
			}

			return v, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func FindTransitGatewayMulticastGroupSourceByThreePartKey(conn *ec2.EC2, multicastDomainID, groupIPAddress, eniID string) (*ec2.TransitGatewayMulticastGroup, error) {
	input := &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"group-ip-address": groupIPAddress,
			"is-group-member":  "false",
			"is-group-source":  "true",
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := FindTransitGatewayMulticastGroups(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output {
		if aws.StringValue(v.NetworkInterfaceId) == eniID {
			// Eventual consistency check.
			if aws.StringValue(v.GroupIpAddress) != groupIPAddress || !aws.BoolValue(v.GroupSource) {
				return nil, &resource.NotFoundError{
					LastRequest: input,
				}
			}

			return v, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func FindTransitGatewayPrefixListReference(conn *ec2.EC2, input *ec2.GetTransitGatewayPrefixListReferencesInput) (*ec2.TransitGatewayPrefixListReference, error) {
	output, err := FindTransitGatewayPrefixListReferences(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayPrefixListReferences(conn *ec2.EC2, input *ec2.GetTransitGatewayPrefixListReferencesInput) ([]*ec2.TransitGatewayPrefixListReference, error) {
	var output []*ec2.TransitGatewayPrefixListReference

	err := conn.GetTransitGatewayPrefixListReferencesPages(input, func(page *ec2.GetTransitGatewayPrefixListReferencesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayPrefixListReferences {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayPrefixListReferenceByTwoPartKey(conn *ec2.EC2, transitGatewayRouteTableID, prefixListID string) (*ec2.TransitGatewayPrefixListReference, error) {
	input := &ec2.GetTransitGatewayPrefixListReferencesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"prefix-list-id": prefixListID,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := FindTransitGatewayPrefixListReference(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.PrefixListId) != prefixListID || aws.StringValue(output.TransitGatewayRouteTableId) != transitGatewayRouteTableID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayRoute(conn *ec2.EC2, transitGatewayRouteTableID, destination string) (*ec2.TransitGatewayRoute, error) {
	input := &ec2.SearchTransitGatewayRoutesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"type": ec2.TransitGatewayRouteTypeStatic,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := conn.SearchTransitGatewayRoutes(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Routes) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, route := range output.Routes {
		if route == nil {
			continue
		}

		if v := aws.StringValue(route.DestinationCidrBlock); verify.CIDRBlocksEqual(v, destination) {
			if state := aws.StringValue(route.State); state == ec2.TransitGatewayRouteStateDeleted {
				return nil, &resource.NotFoundError{
					Message:     state,
					LastRequest: input,
				}
			}

			route.DestinationCidrBlock = aws.String(verify.CanonicalCIDRBlock(v))

			return route, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindTransitGatewayRouteTables(conn *ec2.EC2, input *ec2.DescribeTransitGatewayRouteTablesInput) ([]*ec2.TransitGatewayRouteTable, error) {
	var output []*ec2.TransitGatewayRouteTable

	err := conn.DescribeTransitGatewayRouteTablesPages(input, func(page *ec2.DescribeTransitGatewayRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayRouteTables {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayRouteTablePropagation(conn *ec2.EC2, transitGatewayRouteTableID string, transitGatewayAttachmentID string) (*ec2.TransitGatewayRouteTablePropagation, error) {
	if transitGatewayRouteTableID == "" {
		return nil, nil
	}

	input := &ec2.GetTransitGatewayRouteTablePropagationsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("transit-gateway-attachment-id"),
				Values: aws.StringSlice([]string{transitGatewayAttachmentID}),
			},
		},
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	var result *ec2.TransitGatewayRouteTablePropagation

	err := conn.GetTransitGatewayRouteTablePropagationsPages(input, func(page *ec2.GetTransitGatewayRouteTablePropagationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, transitGatewayRouteTablePropagation := range page.TransitGatewayRouteTablePropagations {
			if transitGatewayRouteTablePropagation == nil {
				continue
			}

			if aws.StringValue(transitGatewayRouteTablePropagation.TransitGatewayAttachmentId) == transitGatewayAttachmentID {
				result = transitGatewayRouteTablePropagation
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func FindTransitGatewayVPCAttachment(conn *ec2.EC2, input *ec2.DescribeTransitGatewayVpcAttachmentsInput) (*ec2.TransitGatewayVpcAttachment, error) {
	output, err := FindTransitGatewayVPCAttachments(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Options == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayVPCAttachments(conn *ec2.EC2, input *ec2.DescribeTransitGatewayVpcAttachmentsInput) ([]*ec2.TransitGatewayVpcAttachment, error) {
	var output []*ec2.TransitGatewayVpcAttachment

	err := conn.DescribeTransitGatewayVpcAttachmentsPages(input, func(page *ec2.DescribeTransitGatewayVpcAttachmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayVpcAttachments {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayVPCAttachmentByID(conn *ec2.EC2, id string) (*ec2.TransitGatewayVpcAttachment, error) {
	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{
		TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayVPCAttachment(conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/vpc/latest/tgw/tgw-vpc-attachments.html#vpc-attachment-lifecycle.
	switch state := aws.StringValue(output.State); state {
	case ec2.TransitGatewayAttachmentStateDeleted,
		ec2.TransitGatewayAttachmentStateFailed,
		ec2.TransitGatewayAttachmentStateRejected:
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}

	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindDHCPOptions(conn *ec2.EC2, input *ec2.DescribeDhcpOptionsInput) (*ec2.DhcpOptions, error) {
	output, err := FindDHCPOptionses(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindDHCPOptionses(conn *ec2.EC2, input *ec2.DescribeDhcpOptionsInput) ([]*ec2.DhcpOptions, error) {
	var output []*ec2.DhcpOptions

	err := conn.DescribeDhcpOptionsPages(input, func(page *ec2.DescribeDhcpOptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DhcpOptions {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidDhcpOptionIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindDHCPOptionsByID(conn *ec2.EC2, id string) (*ec2.DhcpOptions, error) {
	input := &ec2.DescribeDhcpOptionsInput{
		DhcpOptionsIds: aws.StringSlice([]string{id}),
	}

	output, err := FindDHCPOptions(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.DhcpOptionsId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindEgressOnlyInternetGateway(conn *ec2.EC2, input *ec2.DescribeEgressOnlyInternetGatewaysInput) (*ec2.EgressOnlyInternetGateway, error) {
	output, err := FindEgressOnlyInternetGateways(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindEgressOnlyInternetGateways(conn *ec2.EC2, input *ec2.DescribeEgressOnlyInternetGatewaysInput) ([]*ec2.EgressOnlyInternetGateway, error) {
	var output []*ec2.EgressOnlyInternetGateway

	err := conn.DescribeEgressOnlyInternetGatewaysPages(input, func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.EgressOnlyInternetGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindEgressOnlyInternetGatewayByID(conn *ec2.EC2, id string) (*ec2.EgressOnlyInternetGateway, error) {
	input := &ec2.DescribeEgressOnlyInternetGatewaysInput{
		EgressOnlyInternetGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindEgressOnlyInternetGateway(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.EgressOnlyInternetGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindFlowLogByID(conn *ec2.EC2, id string) (*ec2.FlowLog, error) {
	input := &ec2.DescribeFlowLogsInput{
		FlowLogIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeFlowLogs(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.FlowLogs) == 0 || output.FlowLogs[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.FlowLogs); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.FlowLogs[0], nil
}

func FindInternetGateway(conn *ec2.EC2, input *ec2.DescribeInternetGatewaysInput) (*ec2.InternetGateway, error) {
	output, err := FindInternetGateways(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindInternetGateways(conn *ec2.EC2, input *ec2.DescribeInternetGatewaysInput) ([]*ec2.InternetGateway, error) {
	var output []*ec2.InternetGateway

	err := conn.DescribeInternetGatewaysPages(input, func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InternetGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidInternetGatewayIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindInternetGatewayByID(conn *ec2.EC2, id string) (*ec2.InternetGateway, error) {
	input := &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindInternetGateway(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.InternetGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindInternetGatewayAttachment(conn *ec2.EC2, internetGatewayID, vpcID string) (*ec2.InternetGatewayAttachment, error) {
	internetGateway, err := FindInternetGatewayByID(conn, internetGatewayID)

	if err != nil {
		return nil, err
	}

	if len(internetGateway.Attachments) == 0 || internetGateway.Attachments[0] == nil {
		return nil, tfresource.NewEmptyResultError(internetGatewayID)
	}

	if count := len(internetGateway.Attachments); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, internetGatewayID)
	}

	attachment := internetGateway.Attachments[0]

	if aws.StringValue(attachment.VpcId) != vpcID {
		return nil, tfresource.NewEmptyResultError(vpcID)
	}

	return attachment, nil
}

func FindKeyPair(conn *ec2.EC2, input *ec2.DescribeKeyPairsInput) (*ec2.KeyPairInfo, error) {
	output, err := FindKeyPairs(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindKeyPairs(conn *ec2.EC2, input *ec2.DescribeKeyPairsInput) ([]*ec2.KeyPairInfo, error) {
	output, err := conn.DescribeKeyPairs(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidKeyPairNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output.KeyPairs, nil
}

func FindKeyPairByName(conn *ec2.EC2, name string) (*ec2.KeyPairInfo, error) {
	input := &ec2.DescribeKeyPairsInput{
		KeyNames: aws.StringSlice([]string{name}),
	}

	output, err := FindKeyPair(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.KeyName) != name {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindManagedPrefixListByID(conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	input := &ec2.DescribeManagedPrefixListsInput{
		PrefixListIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeManagedPrefixLists(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidPrefixListIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PrefixLists) == 0 || output.PrefixLists[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.PrefixLists); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	prefixList := output.PrefixLists[0]

	if state := aws.StringValue(prefixList.State); state == ec2.PrefixListStateDeleteComplete {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return prefixList, nil
}

func FindManagedPrefixListEntriesByID(conn *ec2.EC2, id string) ([]*ec2.PrefixListEntry, error) {
	input := &ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: aws.String(id),
	}

	var prefixListEntries []*ec2.PrefixListEntry

	err := conn.GetManagedPrefixListEntriesPages(input, func(page *ec2.GetManagedPrefixListEntriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, entry := range page.Entries {
			if entry == nil {
				continue
			}

			prefixListEntries = append(prefixListEntries, entry)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidPrefixListIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return prefixListEntries, nil
}

func FindManagedPrefixListEntryByIDAndCIDR(conn *ec2.EC2, id, cidr string) (*ec2.PrefixListEntry, error) {
	prefixListEntries, err := FindManagedPrefixListEntriesByID(conn, id)

	if err != nil {
		return nil, err
	}

	for _, entry := range prefixListEntries {
		if aws.StringValue(entry.Cidr) == cidr {
			return entry, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindNATGateway(conn *ec2.EC2, input *ec2.DescribeNatGatewaysInput) (*ec2.NatGateway, error) {
	output, err := FindNATGateways(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindNATGateways(conn *ec2.EC2, input *ec2.DescribeNatGatewaysInput) ([]*ec2.NatGateway, error) {
	var output []*ec2.NatGateway

	err := conn.DescribeNatGatewaysPages(input, func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NatGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeNatGatewayNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNATGatewayByID(conn *ec2.EC2, id string) (*ec2.NatGateway, error) {
	input := &ec2.DescribeNatGatewaysInput{
		NatGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNATGateway(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.NatGatewayStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.NatGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindPlacementGroupByName(conn *ec2.EC2, name string) (*ec2.PlacementGroup, error) {
	input := &ec2.DescribePlacementGroupsInput{
		GroupNames: aws.StringSlice([]string{name}),
	}

	output, err := conn.DescribePlacementGroups(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidPlacementGroupUnknown) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PlacementGroups) == 0 || output.PlacementGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.PlacementGroups); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	placementGroup := output.PlacementGroups[0]

	if state := aws.StringValue(placementGroup.State); state == ec2.PlacementGroupStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return placementGroup, nil
}

func FindVPCEndpointConnectionByServiceIDAndVPCEndpointID(conn *ec2.EC2, serviceID, vpcEndpointID string) (*ec2.VpcEndpointConnection, error) {
	input := &ec2.DescribeVpcEndpointConnectionsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"service-id": serviceID,
			// "InvalidFilter: The filter vpc-endpoint-id  is invalid"
			// "vpc-endpoint-id ": vpcEndpointID,
		}),
	}

	var output *ec2.VpcEndpointConnection

	err := conn.DescribeVpcEndpointConnectionsPages(input, func(page *ec2.DescribeVpcEndpointConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VpcEndpointConnections {
			if aws.StringValue(v.VpcEndpointId) == vpcEndpointID {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if vpcEndpointState := aws.StringValue(output.VpcEndpointState); vpcEndpointState == VpcEndpointStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     vpcEndpointState,
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSnapshotById(conn *ec2.EC2, name string) (*ec2.Snapshot, error) {
	input := &ec2.DescribeSnapshotsInput{
		SnapshotIds: aws.StringSlice([]string{name}),
	}

	output, err := conn.DescribeSnapshots(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSnapshotNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Snapshots) == 0 || output.Snapshots[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Snapshots); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Snapshots[0], nil
}

func FindSnapshotTierStatusById(conn *ec2.EC2, id string) (*ec2.SnapshotTierStatus, error) {
	filters := map[string]string{
		"snapshot-id": id,
	}

	input := &ec2.DescribeSnapshotTierStatusInput{
		Filters: BuildAttributeFilterList(filters),
	}

	output, err := conn.DescribeSnapshotTierStatus(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSnapshotNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.SnapshotTierStatuses) == 0 || output.SnapshotTierStatuses[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.SnapshotTierStatuses); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.SnapshotTierStatuses[0], nil
}
