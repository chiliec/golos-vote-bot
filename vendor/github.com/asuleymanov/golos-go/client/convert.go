package client

import (
	"strconv"
	"strings"
)

func (api *Client) SteemPerMvest() (float64, error) {
	dgp, errdgp := api.Rpc.Database.GetDynamicGlobalProperties()
	if errdgp != nil {
		return 0, errdgp
	}

	tvfs, errtvfs := strconv.ParseFloat(strings.Split(dgp.TotalVersingFundSteem, " ")[0], 64)
	if errtvfs != nil {
		return 0, errtvfs
	}

	tvs, errtvs := strconv.ParseFloat(strings.Split(dgp.TotalVestingShares, " ")[0], 64)
	if errtvs != nil {
		return 0, errtvs
	}

	spmtmp := (tvfs / tvs) * 1000000
	str := strconv.FormatFloat(spmtmp, 'f', 3, 64)

	spm, errspm := strconv.ParseFloat(str, 64)
	if errspm != nil {
		return 0, errspm
	}

	return spm, nil
}
