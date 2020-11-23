/**
 * Created by zc on 2020/10/4.
 */
package route

const (
	PathDefineRepo  = "repo"
	PathDefineBuild = "build"
	PathDefineStage = "stage"
	PathDefineStep  = "step"
)

const (
	PathParamRepo  = "{" + PathDefineRepo + "}"
	PathParamBuild = "{" + PathDefineBuild + "}"
	PathParamStage = "{" + PathDefineStage + "}"
	PathParamStep  = "{" + PathDefineStep + "}"
)

const (
	APIRouteRepo   = "/repo"
	APIRouteBuild  = APIRouteRepoPath + "/build"
	APIRouteStream = "/stream"
)

const (
	APIRouteRepoPath     = APIRouteRepo + "/" + PathParamRepo
	APIRouteRepoInfoPath = APIRouteRepo + "/info"
	APIRouteBuildPath    = APIRouteBuild + "/" + PathParamBuild
	APIRouteStepLogPath  = APIRouteBuildPath + "/log/" + PathParamStage + "/" + PathParamStep
)
