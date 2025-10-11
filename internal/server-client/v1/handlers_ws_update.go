package v1

import "github.com/labstack/echo/v4"

func (h Handlers) GetWs(eCtx echo.Context) error {
	if err := h.wsUpdateHttpReqUseCase.Handle(eCtx); err != nil {
		return err
	}
	return nil
}
