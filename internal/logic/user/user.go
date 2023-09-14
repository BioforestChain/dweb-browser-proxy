package user

import (
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/crypto/gmd5"
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
	sUser struct{}
)

func init() {
	service.RegisterUser(New())
}
func New() service.IUser {
	return &sUser{}
}

func (s *sUser) IsDomainExist(ctx context.Context, in model.CheckUrlInput) bool {
	count, err := dao.App.Ctx(ctx).Fields("id").Where(do.App{
		Domain: in.Host,
	}).Count()
	if err != nil {
		return false
	}
	return count > 0
}
func (s *sUser) IsDeviceExist(ctx context.Context, in model.CheckDeviceInput) bool {
	count, err := dao.Device.Ctx(ctx).Fields("id").Where(do.Device{
		Identification: in.DeviceIdentification,
	}).Count()
	if err != nil {
		return false
	}
	return count > 0
}

// Create creates user account.
func (s *sUser) Create(ctx context.Context, in model.UserCreateInput) (entity *v1.ClientRegRes, err error) {
	//domain

	//设备标识用户公钥生成
	md5DeviceIdentification, _ := s.GenerateMD5ByDeviceIdentification(in.PublicKey)
	var (
		available bool
		getUserId uint32
	)
	availableIsIdentification, err := s.IsIdentificationAvailable(ctx, md5DeviceIdentification)
	if err != nil {
		return nil, err
	}
	//TODO 暂定 没有用户名用设备标识填充
	if in.Name == "" {
		in.Name = md5DeviceIdentification
	}
	// Name checks.
	available, err = s.IsNameAvailable(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	if !available {
		//已存在的查出用户id
		getUserId, err = s.GetUserId(ctx, in.Name)
		if err != nil {
			return nil, err
		}
		//return gerror.Newf(`Name "%s" is already token by others`, in.Name)
	}
	nowTimestamp := time.Now().Unix()
	var (
		result  sql.Result
		reqData model.DataToDevice
	)
	err = dao.ProxyServerUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		reqData.Identification = md5DeviceIdentification
		reqData.SrcIdentification = in.Identification
		reqData.Timestamp = nowTimestamp
		reqData.Remark = in.Remark
		if getUserId > 0 {
			reqData.UserId = getUserId
			if availableIsIdentification {
				result, err = s.InsertDevice(ctx, tx, reqData)
				//return nil, gerror.Newf(`DeviceIdentification "%s" is already token by others`, in.Identification)
			}
			if err != nil {
				return err
			}
		} else {
			//
			rootDomainName, _ := g.Cfg().Get(ctx, "root_domain.name")
			domain := in.Name + "." + rootDomainName.String()
			result, err = dao.ProxyServerUser.Ctx(ctx).Data(do.User{
				Name:      in.Name,
				PublicKey: in.PublicKey,
				Domain:    domain,
				Timestamp: nowTimestamp,
				Remark:    in.Remark,
			}).Insert()
			if err != nil {
				return err
			}
			// 同时入库device表
			getUserId, err := result.LastInsertId()
			if err != nil {
				return err
			}
			reqData.UserId = uint32(getUserId)
			if availableIsIdentification {
				result, err = s.InsertDevice(ctx, tx, reqData)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &v1.ClientRegRes{
		md5DeviceIdentification,
		reqData.UserId}, err
}

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
func (s *sUser) GetUserList(ctx context.Context, in model.UserQueryInput) (entities []*do.User, total int, err error) {
	//condition := g.Map{
	//	"name like ?": "%" + in.Name + "%",
	//}
	all, total, err := dao.ProxyServerUser.Ctx(ctx).Offset(in.Offset).Limit(in.Limit).AllAndCount(true)
	if err != nil {
		return nil, 0, err
	}
	if err = all.Structs(&entities); err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}
	return entities, total, err
}
func (s *sUser) GetDomainInfo(ctx context.Context, in model.AppQueryInput) (entities *v1.ClientQueryRes, err error) {
	//app 域名全局唯一
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
	userInfo, err := dao.ProxyServerUser.Ctx(ctx).Fields("domain").Where(conUser).One()
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

// GenerateMD5ByDeviceIdentification
//
//	@Description: 生成md5
//	@receiver s
//	@param identification
//	@return string
//	@return error
func (s *sUser) GenerateMD5ByDeviceIdentification(identification string) (string, error) {
	str, _ := gmd5.EncryptBytes([]byte(identification))
	return str, nil
}

// IsIdentificationAvailable
//
//	@Description: 设备表是否有重复数据
//	@receiver s
//	@param ctx
//	@param identification
//	@return bool
//	@return error
func (s *sUser) IsIdentificationAvailable(ctx context.Context, identification string) (bool, error) {
	count, err := dao.Device.Ctx(ctx).Where(do.Device{
		Identification: identification,
	}).Count()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// IsNameAvailable checks and returns given Name is available for signing up.
func (s *sUser) IsNameAvailable(ctx context.Context, Name string) (bool, error) {
	count, err := dao.ProxyServerUser.Ctx(ctx).Where(do.User{
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
	userId, err := dao.ProxyServerUser.Ctx(ctx).Fields("id").Where(do.User{
		Name: Name,
	}).Value()
	if err != nil {
		return 0, err
	}
	return userId.Uint32(), nil
}
func (s *sUser) GetDeviceId(ctx context.Context, DeviceIdentification string) (int, error) {
	md5DeviceIdentification, _ := s.GenerateMD5ByDeviceIdentification(DeviceIdentification)
	deviceId, err := dao.Device.Ctx(ctx).Fields("id").Where(do.Device{
		Identification: md5DeviceIdentification,
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

// CreateAppInfo
//
//	@Description: 注册域名
//	@receiver s
//	@param ctx
//	@param Name
//	@return int
//	@return error
func (s *sUser) CreateAppInfo(ctx context.Context, in model.UserAppInfoCreateInput) (err error) {
	var (
		//available   bool
		getUserId   uint32
		getDeviceId int
	)
	//// domain checks.
	//available, err = s.IsDomainAvailable(ctx, in.Domain)
	//if err != nil {
	//	return err
	//}
	//if !available {
	//	return gerror.Newf(`Domain "%s" is already token by others`, in.Domain)
	//}
	getUserId, err = s.GetUserId(ctx, in.UserName)
	if err != nil {
		return err
	}
	if getUserId == 0 {
		return gerror.Newf(`UserName "%s" is not registered!`, in.UserName)
	}
	//设备id
	//getDeviceId, err = s.GetDeviceId(ctx, in.DeviceIdentification)
	getDeviceId, err = s.GetDeviceId(ctx, in.PublicKey)
	if err != nil {
		return err
	}
	if getDeviceId == 0 {
		return gerror.Newf(`The DeviceIdIdentification "%s" is not registered!`, in.DeviceIdentification)
	}
	nowTimestamp := time.Now().Unix()
	return dao.App.Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
		_, err = dao.App.Ctx(ctx).Data(do.App{
			UserId:         getUserId,
			DeviceId:       getDeviceId,
			Name:           in.AppName,
			Identification: in.AppIdentification,
			//Domain:         in.Domain,
			Timestamp: nowTimestamp,
			Remark:    in.Remark,
		}).Insert()
		if err != nil {
			return err
		}
		return nil
	})
}
