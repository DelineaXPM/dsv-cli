module thy

replace github.com/spf13/viper => github.com/thycotic-rd/viper v1.7.1

require (
	github.com/Azure/go-autorest/autorest v0.11.21
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.8
	github.com/apex/log v1.1.0
	github.com/atotto/clipboard v0.1.1
	github.com/aws/aws-sdk-go v1.40.37
	github.com/danieljoos/wincred v1.0.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fatih/color v1.7.0
	github.com/gobuffalo/uuid v2.0.5+incompatible
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20190818114111-108c894c2c0e
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/jarcoal/httpmock v1.0.4
	github.com/maxbrunsfeld/counterfeiter/v6 v6.3.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/posener/complete v1.1.1
	github.com/savaki/jq v0.0.0-20161209013833-0e6baecebbf8
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.7.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/thycotic-rd/cli v1.0.1-0.20190221164533-c25d734d6e3d
	github.com/tidwall/pretty v0.0.0-20180105212114-65a9db5fad51
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20210423082822-04245dca01da
	google.golang.org/api v0.13.0
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

go 1.16
