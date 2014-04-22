package core

// NtType declares notification types shared between the various packages.
const (
	NtRefreshAll      NtType = "refreshall"
	NtAccountRefresh  NtType = "accountrefresh"
	NtAccountFeedDone NtType = "accountfeeddone"
)

// NtTypes returns all official NtTypes used in the application.
func NtTypes() []NtType {
	ntTypes := []NtType{}
	ntTypes = append(ntTypes, NtRefreshAll)
	ntTypes = append(ntTypes, NtAccountRefresh)
	ntTypes = append(ntTypes, NtAccountFeedDone)
	return ntTypes
}
