package console

import (
	"fmt"

	"github.com/eiannone/keyboard"
)

func PaginateRaw[T any](paginationEnabled bool, x []T, skip uint, size uint) []T {
	if !paginationEnabled {
		return x
	}

	if skip > uint(len(x)) {
		skip = uint(len(x))
	}

	end := skip + size
	if end > uint(len(x)) {
		end = uint(len(x))
	}

	return x[skip:end]
}

type Pagination struct {
	Enabled     bool
	ItemCount   uint
	PageSize    uint
	CurrentPage uint
}

func (p *Pagination) NextPage() {
	if p.PageSize*p.CurrentPage < p.ItemCount {
		p.CurrentPage += 1
	}
}

func (p *Pagination) FirstItemIndex() uint {
	return p.PageSize*(p.CurrentPage-1) + 1
}

func (p *Pagination) LastItemIndex() uint {
	return p.PageSize*(p.CurrentPage-1) + min(p.ItemCount-p.PageSize*(p.CurrentPage-1), p.PageSize)
}

func (p *Pagination) PrevPage() {
	if p.CurrentPage > 1 {
		p.CurrentPage -= 1
	}
}

func (p *Pagination) PaginationMenu() bool {
	fmt.Printf("\n--- Page: %02d [%d-%d/%d] MENU: <- Prev page | -> Next page | ESC Go back\n", p.CurrentPage, p.FirstItemIndex(), p.LastItemIndex(), p.ItemCount)

	keysEvents, err := keyboard.GetKeys(1)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()
	event := <-keysEvents

	switch event.Key {
	case keyboard.KeyArrowLeft:
		p.PrevPage()
	case keyboard.KeyArrowRight:
		p.NextPage()
	case keyboard.KeyCtrlC:
		return false
	case keyboard.KeyEsc:
		return false
	}

	return true
}

func Paginate[T any](pagination Pagination, x []T) []T {
	if !pagination.Enabled {
		return x
	}

	return PaginateRaw(pagination.Enabled, x, pagination.PageSize*(pagination.CurrentPage-1), pagination.PageSize)
}
