package admin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/pkg/utils"
	mapset "github.com/deckarep/golang-set"
	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

func UserUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(8 << 20)
	file, header, err := r.FormFile("file")
	if err != nil || !strings.Contains(header.Filename, ".xlsx") || !strings.Contains(header.Filename, ".xls") {
		RespError(w, RespInternalErr, "文件解析失败:仅支持xlsx或xls文件")
		return
	}
	defer file.Close()

	// go/path-injection
	// base.Cfg.FilesPath 可以直接对外访问，不能上传文件到此
	fileName := path.Join(os.TempDir(), utils.RandomRunes(10))
	newFile, err := os.Create(fileName)
	if err != nil {
		RespError(w, RespInternalErr, "创建文件失败:", err)
		return
	}
	defer newFile.Close()

	io.Copy(newFile, file)
	if err = UploadUser(newFile.Name()); err != nil {
		RespError(w, RespInternalErr, err)
		os.Remove(fileName)
		return
	}
	os.Remove(fileName)
	RespSucess(w, "批量添加成功")
}

func UploadUser(file string) error {
	f, err := excelize.OpenFile(file)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return err
	}
	if rows[0][0] != "id" || rows[0][1] != "username" || rows[0][2] != "nickname" || rows[0][3] != "email" || rows[0][4] != "pin_code" || rows[0][5] != "limittime" || rows[0][6] != "otp_secret" || rows[0][7] != "disable_otp" || rows[0][8] != "groups" || rows[0][9] != "status" || rows[0][10] != "send_email" {
		return fmt.Errorf("批量添加失败，表格格式不正确")
	}
	var k []interface{}
	for _, v := range dbdata.GetGroupNames() {
		k = append(k, v)
	}
	for index, row := range rows {
		if index == 0 {
			continue
		}
		id, _ := strconv.Atoi(row[0])
		if len(row[4]) < 6 {
			row[4] = utils.RandomRunes(8)
		}
		limittime, _ := time.ParseInLocation("2006-01-02 15:04:05", row[5], time.Local)
		disableOtp, _ := strconv.ParseBool(row[7])
		var group []string
		if row[8] == "" {
			return fmt.Errorf("第%d行数据错误，用户组不允许为空", index)
		}
		for _, v := range strings.Split(row[8], ",") {
			if s := mapset.NewSetFromSlice(k); s.Contains(v) {
				group = append(group, v)
			} else {
				return fmt.Errorf("用户组【%s】不存在,请检查第%d行数据", v, index)
			}
		}
		status := cast.ToInt8(row[9])
		sendmail, _ := strconv.ParseBool(row[10])
		// createdAt, _ := time.ParseInLocation("2006-01-02 15:04:05", row[11], time.Local)
		// updatedAt, _ := time.ParseInLocation("2006-01-02 15:04:05", row[12], time.Local)
		user := &dbdata.User{
			Id:         id,
			Username:   row[1],
			Nickname:   row[2],
			Email:      row[3],
			PinCode:    row[4],
			LimitTime:  &limittime,
			OtpSecret:  row[6],
			DisableOtp: disableOtp,
			Groups:     group,
			Status:     status,
			SendEmail:  sendmail,
			// CreatedAt:  createdAt,
			// UpdatedAt:  updatedAt,
		}
		if err := dbdata.AddBatch(user); err != nil {
			return fmt.Errorf("请检查第%d行数据是否导入有重复用户", index)
		}
		if user.SendEmail {
			if err := userAccountMail(user); err != nil {
				return err
			}
		}
	}
	return nil
}
