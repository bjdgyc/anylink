package admin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

func UploadUser(file string) error {
	user := &dbdata.User{}
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
	for index, row := range rows[:] {
		if index == 0 {
			continue
		}
		id, _ := strconv.Atoi(row[0])
		if len(row[4]) < 6 {
			row[4] = utils.RandomRunes(8)
		}
		limittime, _ := time.Parse("2006-01-02 15:04:05", row[5])
		disableOtp, _ := strconv.ParseBool(row[7])
		group := []string{row[8]}
		status := cast.ToInt8(row[9])
		sendmail, _ := strconv.ParseBool(row[10])
		// createdAt, _ := time.Parse("2006-01-02 03:04:05", row[11])
		// updatedAt, _ := time.Parse("2006-01-02 03:04:05", row[12])
		user = &dbdata.User{
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
			return fmt.Errorf("请检查是否导入有重复用户")
		}
	}
	if user.SendEmail {
		err := userAccountMail(user)
		return err
	}
	return nil
}
