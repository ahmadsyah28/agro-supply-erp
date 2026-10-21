package supplier

import (
	"errors"
	"strings"
	"time"
)

// Supplier merepresentasikan data supplier/pemasok di database
type Supplier struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	ContactName string    `json:"contact_name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateSupplierRequest merepresentasikan payload saat mendaftarkan supplier baru
type CreateSupplierRequest struct {
	Name        string `json:"name"`
	ContactName string `json:"contact_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
}

// Validate memeriksa kelayakan input saat pendaftaran supplier
func (r *CreateSupplierRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.ContactName = strings.TrimSpace(r.ContactName)
	r.Email = strings.TrimSpace(r.Email)
	r.Phone = strings.TrimSpace(r.Phone)
	r.Address = strings.TrimSpace(r.Address)

	if r.Name == "" {
		return errors.New("nama supplier wajib diisi")
	}
	if len(r.Name) > 255 {
		return errors.New("nama supplier maksimal 255 karakter")
	}
	if len(r.ContactName) > 255 {
		return errors.New("nama kontak PIC maksimal 255 karakter")
	}
	if len(r.Email) > 255 {
		return errors.New("email maksimal 255 karakter")
	}
	if len(r.Phone) > 64 {
		return errors.New("nomor telepon maksimal 64 karakter")
	}

	return nil
}

// UpdateSupplierRequest merepresentasikan payload saat memperbarui data supplier
type UpdateSupplierRequest struct {
	Name        string `json:"name"`
	ContactName string `json:"contact_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	IsActive    *bool  `json:"is_active"`
}

// Validate memeriksa kelayakan input saat update supplier
func (r *UpdateSupplierRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.ContactName = strings.TrimSpace(r.ContactName)
	r.Email = strings.TrimSpace(r.Email)
	r.Phone = strings.TrimSpace(r.Phone)
	r.Address = strings.TrimSpace(r.Address)

	if r.Name == "" {
		return errors.New("nama supplier wajib diisi")
	}
	if len(r.Name) > 255 {
		return errors.New("nama supplier maksimal 255 karakter")
	}
	if len(r.ContactName) > 255 {
		return errors.New("nama kontak PIC maksimal 255 karakter")
	}
	if len(r.Email) > 255 {
		return errors.New("email maksimal 255 karakter")
	}
	if len(r.Phone) > 64 {
		return errors.New("nomor telepon maksimal 64 karakter")
	}

	return nil
}
