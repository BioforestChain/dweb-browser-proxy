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

// Create creates user account.
func (s *sUser) Create(ctx context.Context, in model.UserCreateInput) (err error) {
	//设备
	md5DeviceIdentification, _ := s.GenerateMD5ByDeviceIdentification(in.Identification)
	var (
		available bool
		getUserId int
	)
	// Identification checks.
	available, err = s.IsIdentificationAvailable(ctx, md5DeviceIdentification)
	if err != nil {
		return err
	}
	if !available {
		return gerror.Newf(`DeviceIdentification "%s" is already token by others`, in.Identification)
	}
	//TODO 暂定
	if in.Name == "" {
		in.Name = md5DeviceIdentification
	}
	// Name checks.
	available, err = s.IsNameAvailable(ctx, in.Name)

	if err != nil {
		return err
	}

	if !available {
		getUserId, err = s.GetUserId(ctx, in.Name)
		if err != nil {
			return err
		}

		//return gerror.Newf(`Name "%s" is already token by others`, in.Name)
	}
	nowTimestamp := time.Now().Unix()

	return dao.ProxyServerUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var (
			result  sql.Result
			err     error
			reqData dataToDevice
		)
		reqData.Identification = md5DeviceIdentification
		reqData.SrcIdentification = in.Identification
		reqData.Timestamp = nowTimestamp
		reqData.Remark = in.Remark
		if getUserId > 0 {
			reqData.UserId = getUserId
			result, err = s.InsertDevice(ctx, tx, reqData)
			if err != nil {
				return err
			}
			return err
		} else {
			result, err = dao.ProxyServerUser.Ctx(ctx).Data(do.User{
				Name:      in.Name,
				PublicKey: in.PublicKey,
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
			reqData.UserId = getUserId
			result, err = s.InsertDevice(ctx, tx, reqData)
			return err
		}
	})
}

type dataToDevice struct {
	UserId interface{} // 用户id
	//Name           interface{} // 名称
	SrcIdentification interface{} // 源设备标识
	Identification    interface{} // md5后设备标识
	Remark            interface{} // 备注信息
	Timestamp         interface{} // 时间戳
}

func (s *sUser) InsertDevice(ctx context.Context, tx gdb.TX, reqData dataToDevice) (result sql.Result, err error) {
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
	condition := g.Map{
		"name like ?": "%" + in.Name + "%",
	}
	all, total, err := dao.ProxyServerUser.Ctx(ctx).Where(condition).Offset(in.Offset).Limit(in.Limit).AllAndCount(true)

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
	//appInfo, err := dao.App.Ctx(ctx).Where(condition).One()
	//fmt.Printf("appInfo: %#v\n", appInfo.Struct()
	//fmt.Printf("appInfo.Map(): %#v\n", appInfo.Map())

	var (
		getUserId int
		//getDeviceId int
	)

	getUserId, err = s.GetUserId(ctx, in.UserName)
	if err != nil {
		return nil, err
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
	one, err := dao.App.Ctx(ctx).Where(condition).One()

	if err != nil {
		return nil, err
	}
	if err = one.Struct(&entities); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
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
func (s *sUser) GetUserId(ctx context.Context, Name string) (int, error) {
	userId, err := dao.ProxyServerUser.Ctx(ctx).Fields("id").Where(do.User{
		Name: Name,
	}).Value()
	if err != nil {
		return 0, err
	}
	return userId.Int(), nil
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

// CreateDomain
//
//	@Description: 注册域名
//	@receiver s
//	@param ctx
//	@param Name
//	@return int
//	@return error
func (s *sUser) CreateDomain(ctx context.Context, in model.UserDomainCreateInput) (err error) {

	var (
		available   bool
		getUserId   int
		getDeviceId int
	)
	// Identification checks.
	available, err = s.IsDomainAvailable(ctx, in.Domain)
	if err != nil {
		return err
	}
	if !available {
		return gerror.Newf(`Domain "%s" is already token by others`, in.Domain)
	}
	getUserId, err = s.GetUserId(ctx, in.UserName)
	if err != nil {
		return err
	}
	//设备id

	getDeviceId, err = s.GetDeviceId(ctx, in.DeviceIdentification)
	if getDeviceId == 0 {
		return gerror.Newf(`The DeviceIdIdentification "%s" is not registered`, in.DeviceIdentification)
	}

	if err != nil {
		return err
	}

	nowTimestamp := time.Now().Unix()

	return dao.App.Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
		if getUserId > 0 {
			_, err = dao.App.Ctx(ctx).Data(do.App{
				UserId:         getUserId,
				DeviceId:       getDeviceId,
				Name:           in.AppName,
				Identification: in.AppIdentification,
				Domain:         in.Domain,
				Timestamp:      nowTimestamp,
				Remark:         in.Remark,
			}).Insert()
			if err != nil {
				return err
			}
			return err
		}
		return nil
	})
}
