package product

import(
	"errors"
	"strings"
	"time"
)

type Product struct{
	ID int64 `json:"id"`
	SKU string `json:"sku"`
	Name string `json:"name"`
	Unit string `json:"unit"`
	IsActive bool `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProductRequest struct{
	Name string `json:"name"`
	Unit string `json:"unit"`
	SKU string `json:"sku"`
}

func (r *CreateProductRequest) Validate()error{
	r.Name = strings.TrimSpace(r.Name)
	r.Unit = strings.TrimSpace(r.Unit)
	r.SKU = strings.TrimSpace(r.SKU)

	if r.Name == ""{
		return errors.New("nama produk wajib diisi")
	}
	if len(r.Name)>255{
		return errors.New("nama produk  maksimal 255 karakter")
	}
	if r.Unit == ""{
		return errors.New("satuan produk wajib diisi (misal: ton, kg, dll)")
	}
	if len(r.Unit)>50{
		return errors.New("satuan produk  maksimal 50 karakter")
	}
	if r.SKU == ""{
		return errors.New("kode unit produk wajib diisi")
	}
	if len(r.SKU)>100{
		return errors.New("kode unit produk maksimal 100 karakter")
	}
	
	return nil
}

type UpdateProductRequest struct{
	Name string `json:"name"`
	Unit string `json:"unit"`
	IsActive *bool `json:"is_active"`
}

func (r *UpdateProductRequest) Validate()error{
	r.Name = strings.TrimSpace(r.Name)
	r.Unit = strings.TrimSpace(r.Unit)
	if r.Name == ""{
		return errors.New("nama produk wajib diisi")
	}
	if len(r.Name)>255{
		return errors.New("nama produk  maksimal 255 karakter")
	}
	if r.Unit == ""{
		return errors.New("satuan produk wajib diisi (misal: ton, kg, dll)")
	}
	if len(r.Unit)>50{
		return errors.New("satuan produk  maksimal 50 karakter")
	}
	return nil
}