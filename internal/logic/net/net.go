package net

import (
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/dao"
	"proxyServer/internal/model"
	"proxyServer/internal/model/do"
	"proxyServer/internal/service"
	"strings"
	"time"
)

type (
	sNet struct{}
)

func init() {
	service.RegisterNet(New())
}
func New() service.INet {
	return &sNet{}
}

// Create net account.
//
//	@Description: create and edit
//	@receiver s
//	@param ctx
//	@param in
//	@return entity
//	@return err
func (s *sNet) CreateNetModule(ctx context.Context, in model.NetModuleCreateInput) (entity *v1.ClientNetModuleDetailRes, err error) {

	var (
		getPriKey    int64
		available    bool
		result       sql.Result
		nowTimestamp = time.Now().Unix()
	)
	// secret checks.
	secret, _ := g.Cfg().Get(ctx, "auth.secret")
	if in.Secret != secret.String() {
		return nil, gerror.Newf(`Sorry, your secret "%s" is wrong yet`, in.Secret)
	}
	rootDomainName, _ := g.Cfg().Get(ctx, "root_domain.name")
	domain := in.Domain + "." + rootDomainName.String()
	if in.RootDomain != rootDomainName.String() {
		return nil, gerror.Newf(`Sorry, your rootDomain "%s" is wrong yet`, in.RootDomain)
	}

	if in.Id > 0 {
		//更新
		err = dao.Net.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			result, err = dao.Net.Ctx(ctx).Data(do.Net{
				Id:        in.Id,
				NetId:     in.NetId,
				Timestamp: nowTimestamp,
				Port:      in.Port,
				Domain:    domain,
			}).Where(g.Map{
				"id =": in.Id,
			}).Save()
			if err != nil {
				return err
			}
			return nil
		})
		getPriKey = in.Id
	} else {
		// IsDomain checks.
		available, err = s.IsDomainExist(ctx, domain)
		if err != nil {
			return nil, err
		}
		if available {
			return nil, gerror.Newf(`Sorry, your domain "%s" has been registered yet`, in.Domain)
		}
		err = dao.Net.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			result, err = dao.Net.Ctx(ctx).Data(do.Net{
				NetId:     in.NetId,
				Timestamp: nowTimestamp,
				Port:      in.Port,
				Domain:    domain,
			}).Insert()
			if err != nil {
				return err
			}
			return nil
		})
		getPriKey, err = result.LastInsertId()
		if err != nil {
			return nil, err
		}
	}

	findOne, err := dao.Net.Ctx(ctx).One(g.Map{
		"id =": getPriKey,
	})
	if err = findOne.Struct(&entity); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	entity.RootDomain = rootDomainName.String()
	entity.PrefixDomain = in.Domain
	return entity, nil
}

// GetNetModuleDetailById
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param in
//	@return entity
//	@return err
func (s *sNet) GetNetModuleDetailById(ctx context.Context, in model.NetModuleDetailInput) (entity *v1.ClientNetModuleDetailRes, err error) {
	findOne, err := dao.Net.Ctx(ctx).One(g.Map{
		"id =": in.Id,
	})
	if err = findOne.Struct(&entity); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	rootDomainName, _ := g.Cfg().Get(ctx, "root_domain.name")
	parts := strings.Split(entity.Domain, ".")
	entity.RootDomain = rootDomainName.String()
	// 取第一个子串
	entity.PrefixDomain = parts[0]
	return entity, err
}

// GetNetModuleList
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param in
//	@return entities
//	@return total
//	@return err
func (s *sNet) GetNetModuleList(ctx context.Context, in model.NetModuleListQueryInput) (entities []*v1.ClientNetModuleDetailRes, total int, err error) {

	condition := g.Map{
		"domain like ?": "%" + in.Domain + "%",
		"net_id like ?": "%" + in.NetId + "%",
		"is_online =":   in.IsOnline,
	}
	all, total, err := dao.Net.Ctx(ctx).Where(condition).Offset(in.Offset).Limit(in.Limit).OrderDesc("id").AllAndCount(true)
	if err != nil {
		return nil, 0, err
	}
	if err = all.Structs(&entities); err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}
	rootDomainName, _ := g.Cfg().Get(ctx, "root_domain.name")
	for key, entity := range entities {
		parts := strings.Split(entity.Domain, ".")
		entities[key].RootDomain = rootDomainName.String()
		entities[key].PrefixDomain = parts[0]
	}
	return entities, total, err
}

//
//// GetDomainInfo
////
////	@Description: app 域名全局唯一
////	@receiver s
////	@param ctx
////	@param in
////	@return entity
////	@return err
//func (s *sNet) GetDomainInfo(ctx context.Context, in model.AppQueryInput) (entity *gvar.Var, err error) {
//	var getNetId uint32
//	if err != nil {
//		return nil, err
//	}
//	if getNetId == 0 {
//		return nil, gerror.Newf(`NetName "%s" is not registered!`, in.NetName)
//	}
//	condition := g.Map{
//		//"name like ?":    "%" + in.AppName + "%",
//		"identification": in.AppId,
//		"user_id":        getUserId,
//		//"device_id":      getDeviceId,
//	}
//	appInfo, err := dao.App.Ctx(ctx).Fields("domain").Where(condition).Value()
//	if err != nil {
//		return nil, err
//	}
//	return appInfo, err
//}

// IsDomainAvailable
//
//	@Description: Net表中域名是否有重复数据
//	@receiver s
//	@param ctx
//	@param identification
//	@return bool
//	@return error
func (s *sNet) IsDomainExist(ctx context.Context, Domain string) (bool, error) {
	count, err := dao.Net.Ctx(ctx).Where(do.Net{
		Domain: Domain,
	}).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

//
//// CreateDomainInfo
////
////	@Description: 注册域名信息
////	@receiver s
////	@param ctx
////	@param Name
////	@return int
////	@return error
//func (s *sNet) CreateDomainInfo(ctx context.Context, in model.UserAppInfoCreateInput) (err error) {
//	//公钥标识-->域名标识生成
//	var (
//		nowTimestamp = time.Now().Unix()
//	)
//	return dao.App.Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
//		_, err = dao.App.Ctx(ctx).Data(do.App{
//			//UserId:         getUserId,
//			UserName:  in.AppName,
//			AppId:     in.AppId,
//			Timestamp: nowTimestamp,
//			PublicKey: in.PublicKey,
//		}).Insert()
//		if err != nil {
//			return err
//		}
//		return nil
//	})
//}
