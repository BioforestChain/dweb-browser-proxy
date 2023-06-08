package user

import (
	"context"
	"fmt"
	"frpConfManagement/internal/dao"
	"frpConfManagement/internal/model"
	"frpConfManagement/internal/model/do"
	"frpConfManagement/internal/service"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"text/template"
	"time"
)

type (
	sUser struct{}
)

func init() {
	service.RegisterUser(New())
}

func New() service.IUser {
	fmt.Println("in ===IUser")
	return &sUser{}
}

// Create creates user account.
func (s *sUser) Create(ctx context.Context, in model.UserCreateInput) (err error) {
	md5Identification, _ := s.GenerateMD5ByIdentification(in.Identification)
	var (
		available bool
	)
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
	return dao.FrpcUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err = dao.FrpcUser.Ctx(ctx).Data(do.FrpcUser{
			Name:           in.Name,
			Identification: md5Identification,
			Timestamp:      nowTimestamp,
			Remark:         in.Remark,
		}).Insert()
		visitorConfig := FrpCVisitor{md5Identification + "_visitor", "visitor", md5Identification, "127.0.0.1", 19999}
		intervieweeConfig := FrpCInterviewee{md5Identification, "127.0.0.1", 19999}
		s.GenerateFrpCVisitorIni(visitorConfig)
		s.GenerateFrpCIntervieweeIni(intervieweeConfig)
		return err
	})
}

func (s *sUser) GenerateMD5ByIdentification(identification string) (string, error) {
	//nowTimestamp := time.Now().Unix()
	//str, _ := gmd5.EncryptBytes([]byte(identification + strconv.Itoa(int(nowTimestamp))))
	str, _ := gmd5.EncryptBytes([]byte(identification))
	//str = strconv.Itoa(int(nowTimestamp)) + "_" + str
	return str, nil
}

func (s *sUser) IsIdentificationAvailable(ctx context.Context, identification string) (bool, error) {
	count, err := dao.FrpcUser.Ctx(ctx).Where(do.FrpcUser{
		Identification: identification,
	}).Count()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// IsNameAvailable checks and returns given Name is available for signing up.
func (s *sUser) IsNameAvailable(ctx context.Context, Name string) (bool, error) {
	count, err := dao.FrpcUser.Ctx(ctx).Where(do.FrpcUser{
		Name: Name,
	}).Count()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

type FrpCVisitor struct {
	VisitorName string
	Role        string
	ServerName  string
	Host        string
	Port        int
}

type FrpCInterviewee struct {
	IntervieweeName string
	Host            string
	Port            int
}

func (s *sUser) GenerateFrpCVisitorIni(visitorConfig FrpCVisitor) error {
	_, filename, _, _ := runtime.Caller(0)
	rootPath := path.Dir(path.Dir(path.Dir(path.Dir(filename))))
	nowTimestamp := strconv.Itoa(int(time.Now().Unix()))
	outPath := rootPath + "/resource/out/" + nowTimestamp + "_" + visitorConfig.ServerName + "_visitor" + ".ini"
	err := RenderIni(rootPath+"/resource/template/frpc_visitor_template.ini", outPath, visitorConfig)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (s *sUser) GenerateFrpCIntervieweeIni(config FrpCInterviewee) error {
	_, filename, _, _ := runtime.Caller(0)
	rootPath := path.Dir(path.Dir(path.Dir(path.Dir(filename))))
	nowTimestamp := strconv.Itoa(int(time.Now().Unix()))
	outPath := rootPath + "/resource/out/" + nowTimestamp + "_" + config.IntervieweeName + "_interviewee" + ".ini"
	err := RenderIni(rootPath+"/resource/template/frpc_interviewee_template.ini", outPath, config)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func RenderIni(filePath string, outputFilePath string, config interface{}) error {
	t, err := template.ParseFiles(filePath)
	if err != nil {
		return err
	}

	file, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = t.Execute(file, config)
	if err != nil {
		return err
	}

	return nil
}
