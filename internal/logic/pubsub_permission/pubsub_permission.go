package pubsub_permission

import (
	"context"
	"database/sql"
	v1 "github.com/BioforestChain/dweb-browser-proxy/api/pubsub_permission/v1"
	"github.com/BioforestChain/dweb-browser-proxy/internal/dao"
	"github.com/BioforestChain/dweb-browser-proxy/internal/model"
	"github.com/BioforestChain/dweb-browser-proxy/internal/model/do"
	entitis "github.com/BioforestChain/dweb-browser-proxy/internal/model/entity"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg/util/strings"
	"github.com/BioforestChain/dweb-browser-proxy/internal/service"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type (
	sPubsubPermission struct{}
)

func init() {
	service.RegisterPermission(New())
}
func New() service.IPermission {
	return &sPubsubPermission{}
}

// Create net account.
//
//	@Description: create and edit
//	@receiver s
//	@param ctx
//	@param in
//	@return entity
//	@return err
func (s *sPubsubPermission) CreatePubsubPermission(ctx context.Context, in model.PubsubPermissionCreateInput) (entity *v1.PubsubPermissionDetailRes, err error) {
	var (
		getPriKey         int64
		result            sql.Result
		queryPid          *gvar.Var
		idSlice           g.Slice
		pubsubUserAclList []*entitis.PubsubUserAcl
		//getPubsubUserAclList gdb.Result
		//getPubsubUserAclList interface{}
	)

	// query
	if queryPid, err = dao.PubsubPermission.Ctx(ctx).Fields("id").Where(g.Map{
		"name =":      in.Name,
		"publisher =": in.XDwebHostMMID,
		"type =":      in.Type,
	}).Value(); err != nil {
		return nil, err
	}
	// update.
	if queryPid.Int() > 0 {
		//del Permission
		if result, err = dao.PubsubPermission.Ctx(ctx).Delete(g.Map{
			"id =": queryPid.Int(),
		}); err != nil {
			return nil, err
		}
		//del PubsubUserAcl
		if result, err = dao.PubsubUserAcl.Ctx(ctx).Delete(g.Map{
			"permission_id =": queryPid.Int(),
		}); err != nil {
			return nil, err
		}

	}
	//save Permission
	err = dao.PubsubPermission.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if result, err = dao.PubsubPermission.Ctx(ctx).Data(do.Permission{
			Name:      in.Name,
			Type:      in.Type,
			Publisher: in.XDwebHostMMID,
		}).Insert(); err != nil {
			return nil
		}
		getPriKey, _ = result.LastInsertId()
		netDomainList := strings.Explode(",", in.NetDomainNames)
		// for Insert PubsubUserAcl
		for _, v := range netDomainList {
			if result, err = dao.PubsubUserAcl.Ctx(ctx).Data(do.PubsubUserAcl{
				PermissionId: getPriKey,
				NetDomain:    v,
			}).Insert(); err != nil {
				return err
			}
			getPubsubUserAclLastInsertId, _ := result.LastInsertId()
			idSlice = append(idSlice, getPubsubUserAclLastInsertId)
		}

		// query PubsubUserAcl
		getPubsubUserAclList, err := dao.PubsubUserAcl.Ctx(ctx).Where("id", idSlice).All()
		if err != nil {
			return err
		}
		if err = getPubsubUserAclList.Structs(&pubsubUserAclList); err != nil && err != sql.ErrNoRows {
			return err
		}
		return nil
	})
	// result PubsubPermissionDetailRes
	entity = new(v1.PubsubPermissionDetailRes)
	entity.Id = int(getPriKey)
	entity.Name = in.Name
	entity.Type = in.Type
	entity.Publisher = in.XDwebHostMMID
	entity.List = pubsubUserAclList
	return entity, nil
}

func (s *sPubsubPermission) IsPubsubPermissionTopicNameExist(ctx context.Context, Name string) (bool, error) {
	count, err := dao.PubsubPermission.Ctx(ctx).Where(do.Permission{
		Name: Name,
	}).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
