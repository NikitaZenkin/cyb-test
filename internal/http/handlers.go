package http

import (
	"encoding/json"
	"net/http"
)

// FQDNsLoad godoc
// @Summary загрузка списка fqdn
// @Tags 	fqdn
// @ID		fqdn-load
// @Produce	json
// @Param input body http.Values true "список fqdn"
// @Success 200 {string} string ""
// @Failure	500	{object} http.Error
// @Router /fqdn/load [post]
func (c *controller) FQDNsLoad(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var params Values
	if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
		ResponseWithError(w, c.logger, http.StatusBadRequest, err)
		return
	}

	err := c.srv.FQDNsLoad(ctx, params)
	if err != nil {
		ResponseWithError(w, c.logger, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// FQDNsGet godoc
// @Summary получение списков fqdn по ip
// @Tags 	fqdn
// @ID		fqdn-get
// @Produce	json
// @Param input body http.Values true "список ip"
// @Success 200 {array} entity.IpFQDNs
// @Failure	500	{object} http.Error
// @Router /fqdn/list [post]
func (c *controller) FQDNsGet(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var params Values
	if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
		ResponseWithError(w, c.logger, http.StatusBadRequest, err)
		return
	}

	result, err := c.srv.FQDNsGet(ctx, params)
	if err != nil {
		ResponseWithError(w, c.logger, http.StatusInternalServerError, err)
		return
	}

	ResponseWithJSON(w, c.logger, http.StatusOK, result)
}
