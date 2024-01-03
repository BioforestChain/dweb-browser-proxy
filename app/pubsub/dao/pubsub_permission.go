// =================================================================================
// This is auto-generated by GoFrame CLI tool only once. Fill this file as you wish.
// =================================================================================

package dao

import (
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/dao/internal"
)

// internalPubsubPermissionDao is internal type for wrapping internal DAO implements.
type internalPubsubPermissionDao = *internal.PubsubPermissionDao

// pubsubPermissionDao is the data access object for table pubsub_permission.
// You can define custom methods on it to extend its functionality as you wish.
type pubsubPermissionDao struct {
	internalPubsubPermissionDao
}

var (
	// PubsubPermission is globally public accessible object for table pubsub_permission operations.
	PubsubPermission = pubsubPermissionDao{
		internal.NewPubsubPermissionDao(),
	}
)

// Fill with you ideas below.
