package handler

import (
	"context"
	"fmt"
	"net/http"
	"sell/api/models"

	"github.com/gin-gonic/gin"
)

// EndSell godoc
// @Router       /end-sell/{id} [PUT]
// @Summary      end sell
// @Description  end sell
// @Tags         sell
// @Accept       json
// @Produce      json
// @Param 		 id path string true "sale_id"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      404  {object}  models.Response
// @Failure      500  {object}  models.Response
func (h Handler) EndSell(c *gin.Context) {
	saleID := c.Param("id")

	baskets, err := h.storage.Basket().GetList(context.Background(), models.GetListRequest{
		Page:   1,
		Limit:  10,
		Search: saleID,
	})
	if err != nil {
		handleResponse(c, "error is while getting  baskets list", http.StatusInternalServerError, err.Error())
		return
	}

	totalPrice := 0

	for _, value := range baskets.Baskets {
		totalPrice += value.Price
	}

	fmt.Println("totalprice", totalPrice)
	id, err := h.storage.Sale().UpdatePrice(context.Background(), totalPrice, saleID)
	if err != nil {
		handleResponse(c, "error is while updating price", http.StatusInternalServerError, err.Error())
		return
	}

	saleDate, err := h.storage.Sale().GetByID(context.Background(), saleID)

	if err != nil {
		handleResponse(c, "error is while updating price", http.StatusInternalServerError, err.Error())
		return
	}

	repoGetList, err := h.storage.Repository().GetList(context.Background(), models.GetListRequest{
		Page:   1,
		Limit:  1000,
		Search: saleDate.BranchID,
	})

	if err != nil {
		handleResponse(c, "error is while updating price", http.StatusInternalServerError, err.Error())
		return
	}

	repoProducts := make(map[string]int)

	for _, v := range repoGetList.Repositories {
		repoProducts[v.ProductID] = v.Count
	}

	for _, v := range baskets.Baskets {

		_, err := h.storage.Repository().UpdateProductQuantity(context.Background(), models.UpdateRepository{
			ProductID: v.ProductID,
			BranchID:  saleDate.BranchID,
			Count:     repoProducts[v.ProductID] - v.Quantity,
		})

		if err != nil {
			handleResponse(c, "error is while updating price", http.StatusInternalServerError, err.Error())
			return
		}

		_, err = h.storage.RTransaction().Create(context.Background(), models.CreateRepositoryTransaction{
			StaffID:                   saleDate.CashierID,
			ProductID:                 v.ProductID,
			RepositoryTransactionType: "minus",
			Price:                     v.Price,
			Quantity:                  v.Quantity,
		})

		if err != nil {
			handleResponse(c, "error is while updating price", http.StatusInternalServerError, err.Error())
			return
		}
	}

	resp, err := h.storage.Sale().GetByID(context.Background(), id)
	if err != nil {
		handleResponse(c, "error is while getting sale by id", http.StatusInternalServerError, err.Error())
		return
	}
	handleResponse(c, "success", http.StatusOK, resp)
}
