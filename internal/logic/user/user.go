package user

import (
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"log"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/dao"
	"proxyServer/internal/model"
	"proxyServer/internal/model/do"
	"proxyServer/internal/service"
	"time"
)

type (
	sUser struct{}
)

func init() {
	service.RegisterUser(New())
}
func New() service.IUser {
	return &sUser{}
}

// IsDomainExist
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param in
//	@return bool
func (s *sUser) IsDomainExist(ctx context.Context, in model.CheckDomainInput) bool {
	count, err := dao.App.Ctx(ctx).Where(do.App{
		Domain: in.Domain,
	}).Count()
	if err != nil {
		return false
	}
	return count > 0
}

// IsDeviceExist
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param in
//	@return bool
func (s *sUser) IsDeviceExist(ctx context.Context, in model.CheckDeviceInput) bool {
	count, err := dao.Device.Ctx(ctx).Where(do.Device{
		Identification: in.DeviceIdentification,
	}).Count()
	if err != nil {
		return false
	}
	return count > 0
}
func (s *sUser) IsUserExist(ctx context.Context, in model.CheckUserInput) bool {
	count, err := dao.User.Ctx(ctx).Where(do.User{
		Identification: in.UserIdentification,
	}).Count()
	if err != nil {
		return false
	}
	return count > 0
}

// Create user account.
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param in
//	@return entity
//	@return err
func (s *sUser) CreateUser(ctx context.Context, in model.UserCreateInput) (entity *v1.ClientRegRes, err error) {
	//用户key标识用户标识生成
	md5PublicKeyIdentification, _ := s.GenerateMD5ByPublicKeyIdentification(in.UserKey)
	var (
		available bool
		getUserId uint32
		result    sql.Result
		reqData   model.DataToDevice
	)
	nowTimestamp := time.Now().Unix()
	//availableIsUserIdentification, err := s.IsUserIdentificationAvailable(ctx, md5PublicKeyIdentification)
	//if err != nil {
	//	return nil, err
	//}
	//TODO 暂定 没有用户名用用户key标识填充
	if in.Name == "" {
		in.Name = md5PublicKeyIdentification
	}
	// UserIdentification checks.
	available, err = s.IsUserIdentificationAvailable(ctx, md5PublicKeyIdentification)
	if err != nil {
		return nil, err
	}
	if !available {
		//已存在的查出用户id
		getUserId, err = s.GetUserIdByIdentification(ctx, md5PublicKeyIdentification)
		if err != nil {
			return nil, err
		}
	}
	err = dao.User.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		reqData.Identification = md5PublicKeyIdentification
		reqData.SrcIdentification = in.Identification
		reqData.Timestamp = nowTimestamp
		reqData.Remark = in.Remark
		if getUserId > 0 {
			reqData.UserId = getUserId
			//if availableIsUserIdentification {
			//	result, err = s.InsertDevice(ctx, tx, reqData)
			//	//return nil, gerror.Newf(`DeviceIdentification "%s" is already token by others`, in.Identification)
			//}
			//if err != nil {
			//	return err
			//}
		} else {
			////domain  = root_domain
			//rootDomainName, _ := g.Cfg().Get(ctx, "root_domain.name")
			//domain := in.Name + "." + rootDomainName.String()
			result, err = dao.User.Ctx(ctx).Data(do.User{
				Name:    in.Name,
				UserKey: in.UserKey,
				//Domain:         domain,
				Timestamp:      nowTimestamp,
				Identification: md5PublicKeyIdentification,
				Remark:         in.Remark,
			}).Insert()
			if err != nil {
				return err
			}
			getUserId, err := result.LastInsertId()
			if err != nil {
				return err
			}
			reqData.UserId = uint32(getUserId)
			//if availableIsUserIdentification {
			//	result, err = s.InsertDevice(ctx, tx, reqData)
			//	if err != nil {
			//		return err
			//	}
			//}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &v1.ClientRegRes{
		md5PublicKeyIdentification,
		reqData.UserId}, err
}

// InsertDevice
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param tx
//	@param reqData
//	@return result
//	@return err
func (s *sUser) InsertDevice(ctx context.Context, tx gdb.TX, reqData model.DataToDevice) (result sql.Result, err error) {
	result, err = g.Model("device").TX(tx).Data(do.Device{
		UserId:            reqData.UserId,
		Identification:    reqData.Identification,
		SrcIdentification: reqData.SrcIdentification,
		Timestamp:         reqData.Timestamp,
		Remark:            reqData.Remark,
	}).Insert()
	return result, err
}

// GetUserList
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param in
//	@return entities
//	@return total
//	@return err
func (s *sUser) GetUserList(ctx context.Context, in model.UserQueryInput) (entities []*do.User, total int, err error) {
	//condition := g.Map{
	//	"name like ?": "%" + in.Name + "%",
	//}
	all, total, err := dao.User.Ctx(ctx).Offset(in.Offset).Limit(in.Limit).AllAndCount(true)
	if err != nil {
		return nil, 0, err
	}
	if err = all.Structs(&entities); err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}
	return entities, total, err
}

// GetDomainInfo
//
//	@Description: user 域名全局唯一
//	@receiver s
//	@param ctx
//	@param in
//	@return entities
//	@return err
func (s *sUser) GetDomainInfo(ctx context.Context, in model.AppQueryInput) (entities *v1.ClientQueryRes, err error) {
	var (
		getUserId  uint32
		entityUser *v1.ClientDomainQueryRes
		//getDeviceId int
	)
	getUserId, err = s.GetUserId(ctx, in.UserName)
	if err != nil {
		return nil, err
	}
	if getUserId == 0 {
		return nil, gerror.Newf(`UserName "%s" is not registered!`, in.UserName)
	}
	//getDeviceId, err = s.GetDeviceId(ctx, in.DeviceIdentification)
	//if err != nil {
	//	return nil, err
	//}
	condition := g.Map{
		//"name like ?":    "%" + in.AppName + "%",
		"identification": in.AppIdentification,
		"user_id":        getUserId,
		//"device_id":      getDeviceId,
	}
	one, err := dao.App.Ctx(ctx).Fields("identification").Where(condition).One()
	if err != nil {
		return nil, err
	}
	conUser := g.Map{
		"id": getUserId,
	}
	userInfo, err := dao.User.Ctx(ctx).Fields("domain").Where(conUser).One()
	if err != nil {
		return nil, err
	}
	if err = one.Struct(&entities); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err = userInfo.Struct(&entityUser); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	entities.Domain = entityUser.Domain
	return entities, err
}

// GenerateMD5ByPublicKeyIdentification
//
//	@Description: 生成md5
//	@receiver s
//	@param identification
//	@return string
//	@return error
func (s *sUser) GenerateMD5ByPublicKeyIdentification(identification string) (string, error) {
	str, _ := gmd5.EncryptBytes([]byte(identification))
	return str, nil
}

// IsUserIdentificationAvailable
//
//	@Description: 用户表是否有重复数据
//	@receiver s
//	@param ctx
//	@param identification
//	@return bool
//	@return error
func (s *sUser) IsUserIdentificationAvailable(ctx context.Context, identification string) (bool, error) {
	count, err := dao.User.Ctx(ctx).Where(do.User{
		Identification: identification,
	}).Count()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// IsNameAvailable
//
//	@Description: IsNameAvailable checks and returns given Name is available for signing up.
//	@receiver s
//	@param ctx
//	@param Name
//	@return bool
//	@return error
func (s *sUser) IsNameAvailable(ctx context.Context, Name string) (bool, error) {
	count, err := dao.User.Ctx(ctx).Where(do.User{
		Name: Name,
	}).Count()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// @Description:
// @receiver s
// @param ctx
// @param Name
// @return bool
// @return error
func (s *sUser) GetUserId(ctx context.Context, Name string) (uint32, error) {
	userId, err := dao.User.Ctx(ctx).Fields("id").Where(do.User{
		Name: Name,
	}).Value()
	if err != nil {
		return 0, err
	}
	return userId.Uint32(), nil
}

// GetUserIdByIdentification
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param Identification
//	@return uint32
//	@return error
func (s *sUser) GetUserIdByIdentification(ctx context.Context, Identification string) (uint32, error) {
	userId, err := dao.User.Ctx(ctx).Fields("id").Where(do.User{
		Identification: Identification,
	}).Value()
	if err != nil {
		return 0, err
	}
	return userId.Uint32(), nil
}

// GetDeviceId
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param DeviceIdentification
//	@return int
//	@return error
func (s *sUser) GetDeviceId(ctx context.Context, DeviceIdentification string) (int, error) {
	md5PublicKeyIdentification, _ := s.GenerateMD5ByPublicKeyIdentification(DeviceIdentification)
	deviceId, err := dao.Device.Ctx(ctx).Fields("id").Where(do.Device{
		Identification: md5PublicKeyIdentification,
	}).Value()
	if err != nil {
		return 0, err
	}
	return deviceId.Int(), nil
}

// IsDomainAvailable
//
//	@Description: App表中域名是否有重复数据
//	@receiver s
//	@param ctx
//	@param identification
//	@return bool
//	@return error
func (s *sUser) IsDomainAvailable(ctx context.Context, domain string) (bool, error) {
	count, err := dao.App.Ctx(ctx).Where(do.App{
		Domain: domain,
	}).Count()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// CreateDomainInfo
//
//	@Description: 注册域名信息
//	@receiver s
//	@param ctx
//	@param Name
//	@return int
//	@return error
func (s *sUser) CreateDomainInfo(ctx context.Context, in model.UserAppInfoCreateInput) (err error) {
	//公钥标识-->域名标识生成
	md5PublicKeyIdentification, _ := s.GenerateMD5ByPublicKeyIdentification(in.PublicKey)
	var (
		getUserId   uint32
		getDeviceId int
	)
	getUserId = in.UserId
	nowTimestamp := time.Now().Unix()
	//TODO 暂定 没有二级域名就用公钥标识填充
	if in.Subdomain == "" {
		in.Subdomain = md5PublicKeyIdentification
	}
	//domain  = root_domain
	rootDomainName, _ := g.Cfg().Get(ctx, "root_domain.name")
	domain := in.Subdomain + "." + rootDomainName.String()
	// Verify domain exists in the database
	valCheckDomain := service.User().IsDomainExist(ctx, model.CheckDomainInput{Domain: domain})
	if valCheckDomain {
		log.Println(gerror.Newf(`Sorry, your domain "%s" has been registered yet`, domain))
		return gerror.Newf(`Sorry, your domain "%s" has been registered yet`, domain)
	}

	return dao.App.Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
		_, err = dao.App.Ctx(ctx).Data(do.App{
			UserId:         getUserId,
			DeviceId:       getDeviceId,
			Domain:         domain,
			Name:           in.AppName,
			Identification: in.AppIdentification,
			Timestamp:      nowTimestamp,
			Remark:         in.Remark,
			PublicKey:      in.PublicKey,
		}).Insert()
		if err != nil {
			return err
		}
		return nil
	})
}

// CreateAppInfo
// @Description: 记录App（模块）安装信息
// @receiver s
// @param ctx
// @param Name
// @return int
// @return error
func (s *sUser) CreateAppInfo(ctx context.Context, in model.UserAppInfoCreateInput) (err error) {
	var (
		getUserId   uint32
		getDeviceId int
	)
	getUserId, err = s.GetUserId(ctx, in.UserName)
	if err != nil {
		return err
	}
	if getUserId == 0 {
		return gerror.Newf(`UserName "%s" is not registered!`, in.UserName)
	}
	nowTimestamp := time.Now().Unix()
	return dao.App.Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
		_, err = dao.App.Ctx(ctx).Data(do.App{
			UserId:         getUserId,
			DeviceId:       getDeviceId,
			Name:           in.AppName,
			Identification: in.AppIdentification,
			Timestamp:      nowTimestamp,
			Remark:         in.Remark,
			IsInstall:      in.IsInstall,
		}).Save()
		if err != nil {
			return err
		}
		return nil
	})
}
