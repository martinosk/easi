package valueobjects

import (
	pl "easi/backend/internal/auth/publishedlanguage"
)

type Permission = pl.Permission

var ErrInvalidPermission = pl.ErrInvalidPermission

var (
	PermComponentsRead   = pl.PermComponentsRead
	PermComponentsWrite  = pl.PermComponentsWrite
	PermComponentsDelete = pl.PermComponentsDelete

	PermViewsRead   = pl.PermViewsRead
	PermViewsWrite  = pl.PermViewsWrite
	PermViewsDelete = pl.PermViewsDelete

	PermCapabilitiesRead   = pl.PermCapabilitiesRead
	PermCapabilitiesWrite  = pl.PermCapabilitiesWrite
	PermCapabilitiesDelete = pl.PermCapabilitiesDelete

	PermDomainsRead   = pl.PermDomainsRead
	PermDomainsWrite  = pl.PermDomainsWrite
	PermDomainsDelete = pl.PermDomainsDelete

	PermUsersRead   = pl.PermUsersRead
	PermUsersManage = pl.PermUsersManage

	PermInvitationsManage = pl.PermInvitationsManage

	PermMetaModelRead  = pl.PermMetaModelRead
	PermMetaModelWrite = pl.PermMetaModelWrite

	PermAuditRead = pl.PermAuditRead

	PermEnterpriseArchRead   = pl.PermEnterpriseArchRead
	PermEnterpriseArchWrite  = pl.PermEnterpriseArchWrite
	PermEnterpriseArchDelete = pl.PermEnterpriseArchDelete

	PermEditGrantsManage = pl.PermEditGrantsManage

	PermValueStreamsRead   = pl.PermValueStreamsRead
	PermValueStreamsWrite  = pl.PermValueStreamsWrite
	PermValueStreamsDelete = pl.PermValueStreamsDelete
)

var PermissionFromString = pl.PermissionFromString

var PermissionsToStrings = pl.PermissionsToStrings
