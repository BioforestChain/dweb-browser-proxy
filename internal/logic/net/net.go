package net

import (
	"context"
	"database/sql"
	v1 "github.com/BioforestChain/dweb-browser-proxy/api/client/v1"
	"github.com/BioforestChain/dweb-browser-proxy/internal/dao"
	"github.com/BioforestChain/dweb-browser-proxy/internal/model"
	"github.com/BioforestChain/dweb-browser-proxy/internal/model/do"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg/rsa"
	stringsHelper "github.com/BioforestChain/dweb-browser-proxy/internal/pkg/util/strings"
	"github.com/BioforestChain/dweb-browser-proxy/internal/service"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"log"
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
		getPriKey              int64
		available              bool
		result                 sql.Result
		nowTimestamp           = time.Now().Unix()
		findOne                gdb.Record
		publicKeyMd5           string
		prefixBroadcastAddress string
	)
	prefixBroadcastAddress = in.PrefixBroadcastAddress

	// secret checks.
	secret, _ := g.Cfg().Get(ctx, "auth.secret")
	if in.Secret != secret.String() {
		return nil, gerror.Newf(`Sorry, your secret "%s" is wrong yet`, in.Secret)
	}
	serverAddr := in.ServerAddr

	if publicKeyMd5, err = gmd5.Encrypt(in.PublicKey); err != nil {
		return nil, err
	}

	// update.
	if in.Id > 0 {
		err = dao.Net.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			result, err = dao.Net.Ctx(ctx).Data(do.Net{
				Id:               in.Id,
				ServerAddr:       serverAddr,
				Port:             in.Port,
				Timestamp:        nowTimestamp,
				BroadcastAddress: in.BroadcastAddress,
				NetId:            in.NetId,
				PublicKey:        in.PublicKey,
				PublicKeyMd5:     publicKeyMd5,
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
		// add.
		// IsServerAddr checks.
		if available, err = s.IsServerAddrExist(ctx, serverAddr); err != nil {
			return nil, err
		}
		if available {
			return nil, gerror.Newf(`Sorry, your server addr "%s" has been registered yet`, in.ServerAddr)
		}

		if available, err = s.IsBroadcastAddressExist(ctx, in.BroadcastAddress, publicKeyMd5); err != nil {
			return nil, err
		}
		if available {
			return nil, gerror.Newf(`Sorry, your broadcast address  "%s" has been registered yet`, in.BroadcastAddress)
		}

		err = dao.Net.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			result, err = dao.Net.Ctx(ctx).Data(do.Net{
				ServerAddr:       serverAddr,
				Port:             in.Port,
				Timestamp:        nowTimestamp,
				BroadcastAddress: in.BroadcastAddress,
				NetId:            in.NetId,
				PublicKey:        in.PublicKey,
				PublicKeyMd5:     publicKeyMd5,
			}).Insert()
			log.Printf("dao Net panic: ", err)
			if err != nil {
				code := gerror.Code(err)
				if code == gcode.CodeDbOperationError {
					return gerror.Newf(`Sorry, your broadcastAddress "%s" has been registered yet`, in.BroadcastAddress)
				} else {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		getPriKey, _ = result.LastInsertId()
	}

	if findOne, err = dao.Net.Ctx(ctx).One(g.Map{"id =": getPriKey}); err != nil {
		return nil, err
	}
	if err = findOne.Struct(&entity); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	entity.Domain = in.ServerAddr

	// 获取 top level domain
	entity.RootDomain = in.RootDomain
	// 用一级域名替换域名得到子串
	entity.PrefixBroadcastAddress = prefixBroadcastAddress
	prvKey, pubKey := s.getPrvPubRSAKey()
	entity.PrivateKey = prvKey
	entity.PublicKey = pubKey
	return entity, nil
}

// getPrvPubRSAKey
//
//	@Description: /Rsa/gen
//	@receiver c
//	@since: time
func (c *sNet) getPrvPubRSAKey() (prvKey, pubKey string) {
	prvKeySrc, pubKeySrc := rsa.GenRsaKey()
	return stringsHelper.BytesToStr(prvKeySrc), stringsHelper.BytesToStr(pubKeySrc)
}

// GetNetModuleDetailById
//
//	@Description:
//	@receiver s
//	@param ctx
//	@param in
//	@return entity
//	@return err
func GetNetPublicKeyByBroAddr(ctx context.Context, in model.NetModulePublicKeyInput) (entity *v1.ClientNetModulePublicKeyRes, err error) {
	findOne, err := dao.Net.Ctx(ctx).Fields("public_key").One(g.Map{
		"broadcast_address =": in.BroadcastAddress,
	})
	if err = findOne.Struct(&entity); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return entity, err
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
	parts := strings.Split(entity.BroadcastAddress, ".")
	// 获取 top level domain
	tld := parts[len(parts)-2] + "." + parts[len(parts)-1]
	entity.RootDomain = tld
	entity.PrefixBroadcastAddress = strings.Replace(entity.BroadcastAddress, "."+tld, "", 1)
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
	for key, entity := range entities {
		parts := strings.Split(entity.BroadcastAddress, ".")
		tld := parts[len(parts)-2] + "." + parts[len(parts)-1]
		entities[key].RootDomain = tld
		entities[key].PrefixBroadcastAddress = strings.Replace(entity.BroadcastAddress, "."+tld, "", 1)
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

// IsServerAddrExist
//
//	@Description: Net表中域名是否有重复数据
//	@receiver s
//	@param ctx
//	@param identification
//	@return bool
//	@return error
func (s *sNet) IsServerAddrExist(ctx context.Context, serverAddr string) (bool, error) {
	count, err := dao.Net.Ctx(ctx).Where(do.Net{
		ServerAddr: serverAddr,
	}).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsBroadcastAddressExist
//
//	@Description: Net表中broadcast_address是否合法存在，无有重复数据
//	@receiver s
//	@param ctx
//	@param identification
//	@return bool
//	@return error
func (s *sNet) IsBroadcastAddressExist(ctx context.Context, broadcastAddress, publicKeyMd5 string) (bool, error) {
	count, err := dao.Net.Ctx(ctx).Where(do.Net{
		BroadcastAddress: broadcastAddress,
		PublicKeyMd5:     publicKeyMd5,
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
