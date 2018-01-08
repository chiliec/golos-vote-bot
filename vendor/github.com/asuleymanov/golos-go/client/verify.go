package client

import (
	// Stdlib
	"fmt"
	"log"
	"strings"
	"strconv"
	"math/big"

	// Vendor
	"github.com/pkg/errors"
)

//We check whether there is a voter on the list of those who have already voted
func (api *Client) Verify_Voter_Weight(author, permlink, voter string, weight int) bool {
	ans, err := api.Rpc.Database.GetActiveVotes(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Voter: "))
		return false
	} else {
		for _, v := range ans {
			if v.Voter == voter && v.Percent == weight {
				return true
			}
		}
		return false
	}
}

func (api *Client) Verify_Voter(author, permlink, voter string) bool {
	ans, err := api.Rpc.Database.GetActiveVotes(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Voter: "))
		return false
	} else {
		for _, v := range ans {
			if v.Voter == voter {
				return true
			}
		}
		return false
	}
}

//We check whether there are voted
func (api *Client) Verify_Votes(author, permlink string) bool {
	ans, err := api.Rpc.Database.GetActiveVotes(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Votes: "))
		return false
	} else {
		if len(ans) > 0 {
			return true
		} else {
			return false
		}
	}
}

func (api *Client) Verify_Comments(author, permlink string) bool {
	ans, err := api.Rpc.Database.GetContentReplies(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Comments: "))
		return false
	} else {
		if len(ans) > 0 {
			return true
		} else {
			return false
		}
	}
}

func (api *Client) Verify_Reblogs(author, permlink, rebloger string) bool {
	ans, err := api.Rpc.Follow.GetRebloggedBy(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Reblogs: "))
		return false
	} else {
		for _, v := range ans {
			if v == rebloger {
				return true
			}
		}
		return false
	}
}

func (api *Client) Verify_Follow(follower, following string) bool {
	ans, err := api.Rpc.Follow.GetFollowing(follower, following, "blog", 1)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Follow: "))
		return false
	} else {
		for _, v := range ans {
			if (v.Follower == follower) && (v.Following == following) {
				return true
			} else {
				return false
			}
		}
		return false
	}
}

func (api *Client) Verify_Post(author, permlink string) bool {
	ans, err := api.Rpc.Database.GetContent(author, permlink)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Post: "))
		return false
	} else {
		if (ans.Author == author) && (ans.Permlink == permlink) {
			return true
		} else {
			return false
		}
		return false
	}
}

func (api *Client) Verify_Delegate_Posting_Key_Sign(username string, arr []string) []string {
	var truearr []string

	props, err := api.Rpc.Database.GetDynamicGlobalProperties()
	if err != nil {
		log.Println(errors.Wrap(err, "Error Get Dynamic Global Properties"))
		return nil
	}
	totalVestingShares, err := strconv.ParseFloat(strings.Split(props.TotalVestingShares, " ")[0], 64)
	if err != nil {
		log.Println(errors.Wrap(err, "Error Parse Total Vesting Shares"))
		return nil
	}
	totalVestingSharesInt, _ := new(big.Float).SetFloat64(totalVestingShares).Int(nil)
	maxVirtualBandwidth := props.MaxVirtualBandwidth.Int

	acc, err := api.Rpc.Database.GetAccounts(arr)
	if err != nil {
		log.Println(errors.Wrapf(err, "Error Verify Delegate Vote Sign: "))
		return nil
	} else {
		for _, val := range acc {
			if !val.CanVote {
				continue
			}
			vestingShares, err := strconv.ParseFloat(strings.Split(val.VestingShares, " ")[0], 64)
			if err != nil {
				log.Println(errors.Wrap(err, "Error Parse Vesting Shares"))
				continue
			}
			if vestingShares < 0 {
				continue
			}
			vestingSharesInt, _ := new(big.Float).SetFloat64(vestingShares).Int(nil)
			firstCondition := new(big.Int).Mul(vestingSharesInt, maxVirtualBandwidth)
			secondCondition := new(big.Int).Mul(val.AverageBandwidth.Int, totalVestingSharesInt)
			log.Printf("Разность: %s", new(big.Int).Sub(firstCondition, secondCondition).String())
			if firstCondition.Cmp(secondCondition) == -1 {
				log.Printf("Аккаунт %s has exceeded maximum allowed bandwidth per vesting share", val.Name)
				continue
			}
			for _, v := range val.Posting.AccountAuths {
				l := strings.Split(strings.Replace(strings.Replace(fmt.Sprintf("%v", v), "[", "", -1), "]", "", -1), " ")[0]
				if l == username {
					truearr = append(truearr, val.Name)
				}
			}
		}
	}

	return truearr
}
