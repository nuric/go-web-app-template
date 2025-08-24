package controllers

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/nuric/go-web-app-template/auth"
	"github.com/nuric/go-web-app-template/models"
	"github.com/nuric/go-web-app-template/utils"
	"gorm.io/gorm"
)

type AccountPage struct {
	BasePage
	User               models.User
	ChangeEmailForm    ChangeEmailForm
	ChangePasswordForm ChangePasswordForm
	UpdateProfileForm  UpdateProfileForm
}

type ChangeEmailForm struct {
	Action     string `schema:"_action"`
	Email      string `schema:"email"`
	EmailError error
	Token      string `schema:"token"`
	TokenError error
	Error      error
}

func (f *ChangeEmailForm) Validate() bool {
	f.EmailError = ValidateEmail(f.Email)
	if f.Action == "change_email" {
		f.TokenError = ValidateToken(f.Token)
	}
	return f.EmailError == nil && f.TokenError == nil
}

type ChangePasswordForm struct {
	CurrentPassword      string `schema:"currentPassword"`
	CurrentPasswordError error
	NewPassword          string `schema:"newPassword"`
	NewPasswordError     error
	ConfirmPassword      string `schema:"confirmPassword"`
	ConfirmPasswordError error
	Error                error
}

func (f *ChangePasswordForm) Validate() bool {
	f.CurrentPasswordError = ValidatePassword(f.CurrentPassword)
	f.NewPasswordError = ValidatePassword(f.NewPassword)
	if f.NewPassword != f.ConfirmPassword {
		f.ConfirmPasswordError = errors.New("passwords do not match")
	}
	return f.CurrentPasswordError == nil && f.NewPasswordError == nil && f.ConfirmPasswordError == nil
}

type UpdateProfileForm struct {
	Name         string `schema:"name"`
	NameError    error
	PictureError error // Used as a placeholder for picture upload errors
	Error        error
}

func (f *UpdateProfileForm) Validate() bool {
	if f.Name == "" {
		f.NameError = errors.New("name is required")
	}
	return f.NameError == nil
}

func (p *AccountPage) Handle(w http.ResponseWriter, r *http.Request) {
	p.User = auth.GetCurrentUser(r)
	// ---------------------------
	if r.Method == http.MethodGet {
		return
	}
	// ---------------------------
	switch r.PostFormValue("_action") {
	case "update_profile":
		f := &p.UpdateProfileForm
		if err := DecodeValidForm(f, r); err != nil {
			f.Error = err
			return
		}
		file, handler, err := r.FormFile("picture")
		if err != nil {
			f.PictureError = err
			return
		}
		defer file.Close()
		if handler.Size > 5*1024*1024 {
			f.PictureError = errors.New("picture size exceeds 5MB limit")
			return
		}
		// We are using a UUID for the filename to avoid collisions
		guid := uuid.New().String()
		fname := fmt.Sprintf("profile/%s%s", guid, filepath.Ext(handler.Filename))
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&p.User).Updates(models.User{Name: f.Name, Picture: "/uploads/" + fname}).Error; err != nil {
				return err
			}
			picFile, err := st.Create(fname)
			if err != nil {
				return err
			}
			defer picFile.Close()
			if _, err := io.Copy(picFile, file); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			slog.Error("could not update user profile", "error", err)
			f.Error = errors.New("could not update user profile")
			return
		}
		p.Flash(r, FlashSuccess, "Your profile has been updated")
		p.redirect = r.URL.Path
	case "request_email_change_token":
		f := &p.ChangeEmailForm
		if err := DecodeValidForm(f, r); err != nil {
			f.Error = err
			return
		}
		if err := sendEmailVerification(p.User.ID, f.Email); err != nil {
			f.Error = err
			return
		}
		// Switch to next action
		f.Action = "change_email"
		p.Flash(r, FlashInfo, "Verification email sent. Please check your inbox.")
	case "change_email":
		f := &p.ChangeEmailForm
		if err := DecodeValidForm(f, r); err != nil {
			f.Error = err
			return
		}
		if err := checkEmailVerification(p.User.ID, f.Email, f.Token); err != nil {
			f.Error = err
			return
		}
		// Update the email of the user
		if err := db.Model(&p.User).Update("email", f.Email).Error; err != nil {
			slog.Error("could not update user email", "error", err)
			f.Error = errors.New("could not change user email")
			return
		}
		p.Flash(r, FlashSuccess, "Your email has been changed")
		p.redirect = r.URL.Path
	case "change_password":
		f := &p.ChangePasswordForm
		if err := DecodeValidForm(f, r); err != nil {
			f.Error = err
			return
		}
		if !utils.VerifyPassword(p.User.Password, f.CurrentPassword) {
			f.CurrentPasswordError = errors.New("please enter your current password")
			return
		}
		hashedPassword := utils.HashPassword(f.NewPassword)
		if err := db.Model(&p.User).Update("password", hashedPassword).Error; err != nil {
			slog.Error("could not change user password", "error", err)
			f.Error = errors.New("could not change user password")
			return
		}
		// Redirect to GET current page
		p.Flash(r, FlashSuccess, "Your password has been changed")
		p.redirect = r.URL.Path
	default:
		p.notFound = true
	}
}
