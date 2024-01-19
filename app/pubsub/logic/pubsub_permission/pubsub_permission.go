package pubsub_permission

import (
	"context"
	"database/sql"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/api/pubsub_permission/v1"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/dao"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/model"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/model/do"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/model/entity"
	"github.com/BioforestChain/dweb-browser-proxy/app/pubsub/service"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/util/strings"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type (
	sPubsubPermission struct{}
)

func init() {
	service.RegisterPubsubPermission(New())
}
func New() service.IPubsubPermission {
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
func (s *sPubsubPermission) CreatePubsubPermission(ctx context.Context, in model.PubsubPermissionCreateInput) (res *v1.PubsubPermissionDetailRes, err error) {
	var (
		getPriKey         int64
		result            sql.Result
		queryPid          *gvar.Var
		idSlice           g.Slice
		pubsubUserAclList []*entity.PubsubUserAcl
		//getPubsubUserAclList gdb.Result
		//getPubsubUserAclList interface{}
	)

	//save Permission
	err = dao.PubsubPermission.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// query
		if queryPid, err = dao.PubsubPermission.Ctx(ctx).Fields("id").Where(g.Map{
			"topic =":     in.Topic,
			"publisher =": in.XDwebHostMMID,
			"type =":      in.Type,
		}).Value(); err != nil {
			return err
		}
		// update.
		if queryPid.Int() > 0 {
			//del Permission
			if result, err = dao.PubsubPermission.Ctx(ctx).Delete(g.Map{
				"id =": queryPid.Int(),
			}); err != nil {
				return err
			}
			//del PubsubUserAcl
			if result, err = dao.PubsubUserAcl.Ctx(ctx).Delete(g.Map{
				"permission_id =": queryPid.Int(),
			}); err != nil {
				return err
			}
		}

		if result, err = dao.PubsubPermission.Ctx(ctx).Data(do.PubsubPermission{
			Topic:     in.Topic,
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
	res = new(v1.PubsubPermissionDetailRes)
	res.Id = int(getPriKey)
	res.Topic = in.Topic
	res.Type = in.Type
	res.Publisher = in.XDwebHostMMID
	res.List = pubsubUserAclList
	return res, nil
}

func (s *sPubsubPermission) IsPubsubPermissionTopicNameExist(ctx context.Context, Topic string) (bool, error) {
	count, err := dao.PubsubPermission.Ctx(ctx).Where(do.PubsubPermission{
		Topic: Topic,
	}).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
