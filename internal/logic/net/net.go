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
//	@Description:
//	@receiver s
//	@param ctx
//	@param in
//	@return entity
//	@return err
func (s *sNet) CreateNetModule(ctx context.Context, in model.NetModuleCreateInputReq) (entity *v1.ClientNetModuleRegRes, err error) {
	var (
		available    bool
		result       sql.Result
		nowTimestamp = time.Now().Unix()
	)
	rootDomainName, _ := g.Cfg().Get(ctx, "root_domain.name")
	domain := in.Domain + "." + rootDomainName.String()
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
			Domain:    domain,
		}).Insert()
		if err != nil {
			return err
		}
		return nil
	})
	getPriKey, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &v1.ClientNetModuleRegRes{
		getPriKey,
		in.NetId}, err
}

//
//// GetUserList
////
////	@Description:
////	@receiver s
////	@param ctx
////	@param in
////	@return entities
////	@return total
////	@return err
//func (s *sNet) GetNetList(ctx context.Context, in model.UserQueryInput) (entities []*do.Net, total int, err error) {
//	all, total, err := dao.Net.Ctx(ctx).Offset(in.Offset).Limit(in.Limit).AllAndCount(true)
//	if err != nil {
//		return nil, 0, err
//	}
//	if err = all.Structs(&entities); err != nil && err != sql.ErrNoRows {
//		return nil, 0, err
//	}
//	return entities, total, err
//}

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
