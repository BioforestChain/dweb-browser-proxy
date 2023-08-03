package user

import (
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"proxyServer/internal/dao"
	"proxyServer/internal/model"
	"proxyServer/internal/model/do"
	"proxyServer/internal/service"
	"time"
)

type (
	sUser struct{}
)

var serverFileIniPath = "/var/opt/"
var serverPort = "10000"

func init() {
	service.RegisterUser(New())
}

func New() service.IUser {
	return &sUser{}
}

// Create creates user account.
func (s *sUser) Create(ctx context.Context, in model.UserCreateInput) (err error) {
	md5Identification, _ := s.GenerateMD5ByIdentification(in.Identification)
	var available bool
	// Identification checks.

	available, err = s.IsIdentificationAvailable(ctx, md5Identification)
	if err != nil {
		return err
	}
	if !available {
		return gerror.Newf(`Identification "%s" is already token by others`, in.Identification)
	}
	if in.Name == "" {
		in.Name = md5Identification
	}
	// Name checks.
	available, err = s.IsNameAvailable(ctx, in.Name)
	if err != nil {
		return err
	}
	if !available {
		return gerror.Newf(`Name "%s" is already token by others`, in.Name)
	}
	nowTimestamp := time.Now().Unix()
	return dao.ProxyServerUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err = dao.ProxyServerUser.Ctx(ctx).Data(do.ProxyServerUser{
			Name:           in.Name,
			Domain:         in.Domain,
			PublicKey:      in.PublicKey,
			Identification: md5Identification,
			Timestamp:      nowTimestamp,
			Remark:         in.Remark,
		}).Insert()
		return err
	})
}
func (s *sUser) GetUserList(ctx context.Context, in model.UserQueryInput) (out []*do.ProxyServerUser, total int, err error) {
	condition := g.Map{
		"domain like ?": "%" + in.Domain + "%",
	}
	all, total, err := dao.ProxyServerUser.Ctx(ctx).Where(condition).Offset(in.Offset).Limit(in.Limit).AllAndCount(true)

	if err != nil {
		return nil, 0, err
	}
	var entities []*do.ProxyServerUser
	if err = all.Structs(&entities); err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}
	return entities, total, err
}

func (s *sUser) GenerateMD5ByIdentification(identification string) (string, error) {
	//nowTimestamp := time.Now().Unix()
	//str, _ := gmd5.EncryptBytes([]byte(identification + strconv.Itoa(int(nowTimestamp))))
	str, _ := gmd5.EncryptBytes([]byte(identification))
	//str = strconv.Itoa(int(nowTimestamp)) + "_" + str
	return str, nil
}

func (s *sUser) IsIdentificationAvailable(ctx context.Context, identification string) (bool, error) {
	count, err := dao.ProxyServerUser.Ctx(ctx).Where(do.ProxyServerUser{
		Identification: identification,
	}).Count()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// IsNameAvailable checks and returns given Name is available for signing up.
func (s *sUser) IsNameAvailable(ctx context.Context, Name string) (bool, error) {
	count, err := dao.ProxyServerUser.Ctx(ctx).Where(do.ProxyServerUser{
		Name: Name,
	}).Count()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
