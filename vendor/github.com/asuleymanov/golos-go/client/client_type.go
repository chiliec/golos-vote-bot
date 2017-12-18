package client

type PC_Options struct {
	Percent uint16
}

type PC_Vote struct {
	Weight int
}

type ArrTransfer struct {
	To      string
	Memo    string
	Ammount string
}

type ArrVote struct {
	User   string
	Weight int
}
