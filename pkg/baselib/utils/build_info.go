package utils

/*
const char* goBaselibBuildTime(void)
{
static const char* pszBuildTime = __DATE__ " " __TIME__ ;
    return pszBuildTime;
}
*/
import "C"

var (
	GitBranch = "not set"
	GitCommit = "not set"
	BuildTime = C.GoString(C.goBaselibBuildTime())
)
