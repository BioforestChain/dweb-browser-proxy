package app

import (
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/database/gdb"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/dao"
	"proxyServer/internal/model"
	"proxyServer/internal/model/do"
	"proxyServer/internal/service"
	"time"
)

type (
	sApp struct{}
)

func init() {
	service.RegisterApp(New())
}
func New() service.IApp {
	return &sApp{}
}

// Create net account.
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param in
//	@return entity
//	@return err
func (s *sApp) CreateAppModule(ctx context.Context, in model.AppModuleCreateInput) (entity *v1.ClientAppModuleRegRes, err error) {
	var (
		//available    bool
		result       sql.Result
		nowTimestamp = time.Now().Unix()
	)

	err = dao.App.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err = dao.App.Ctx(ctx).Data(do.App{
			AppId:     in.AppId,
			NetId:     in.NetId,
			Timestamp: nowTimestamp,
			UserName:  in.UserName,
			AppName:   in.AppName,
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
	return &v1.ClientAppModuleRegRes{
		getPriKey,
		in.NetId, in.AppId}, err
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
//func (s *sApp) GetNetList(ctx context.Context, in model.UserQueryInput) (entities []*do.App, total int, err error) {
//	all, total, err := dao.App.Ctx(ctx).Offset(in.Offset).Limit(in.Limit).AllAndCount(true)
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
//func (s *sApp) GetDomainInfo(ctx context.Context, in model.AppQueryInput) (entity *gvar.Var, err error) {
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

//
//// CreateDomainInfo
////
////	@Description: 注册域名信息
////	@receiver s
////	@param ctx
////	@param Name
////	@return int
////	@return error
//func (s *sApp) CreateDomainInfo(ctx context.Context, in model.UserAppInfoCreateInput) (err error) {
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

// CreateAppInfo
// @Description: 记录App（模块）安装信息
// @receiver s
// @param ctx
// @param Name
// @return int
// @return error
func (s *sApp) CreateAppInfo(ctx context.Context, in model.AppModuleInfoCreateInput) (err error) {
	var (
		nowTimestamp = time.Now().Unix()
	)
	//getUserId, err = s.GetUserId(ctx, in.UserName)
	if err != nil {
		return err
	}
	return dao.App.Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
		_, err = dao.App.Ctx(ctx).Data(do.App{
			AppId:     in.AppId,
			NetId:     in.NetId,
			UserName:  in.UserName,
			AppName:   in.AppName,
			Timestamp: nowTimestamp,
			PublicKey: in.PublicKey,
			IsInstall: in.IsInstall,
			IsOnline:  in.IsOnline,
		}).Save()
		if err != nil {
			return err
		}
		return nil
	})
}
