package users

import (
    "context"
    i2ae4187f7daee263371cb1c977df639813ab50ffa529013b7437480d1ec0158f "github.com/microsoft/kiota-abstractions-go"
    ia572726a95efa92ddd544552cd950653dc691023836923576b2f4bf716cf204a "github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
)

// ItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilder provides operations to call the disableSmsSignIn method.
type ItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilder struct {
    // Path parameters for the request
    pathParameters map[string]string
    // The request adapter to use to execute the requests.
    requestAdapter i2ae4187f7daee263371cb1c977df639813ab50ffa529013b7437480d1ec0158f.RequestAdapter
    // Url template to use to build the URL for the current request builder
    urlTemplate string
}
// ItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilderPostRequestConfiguration configuration for the request such as headers, query parameters, and middleware options.
type ItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilderPostRequestConfiguration struct {
    // Request headers
    Headers *i2ae4187f7daee263371cb1c977df639813ab50ffa529013b7437480d1ec0158f.RequestHeaders
    // Request options
    Options []i2ae4187f7daee263371cb1c977df639813ab50ffa529013b7437480d1ec0158f.RequestOption
}
// NewItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilderInternal instantiates a new DisableSmsSignInRequestBuilder and sets the default values.
func NewItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilderInternal(pathParameters map[string]string, requestAdapter i2ae4187f7daee263371cb1c977df639813ab50ffa529013b7437480d1ec0158f.RequestAdapter)(*ItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilder) {
    m := &ItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilder{
    }
    m.urlTemplate = "{+baseurl}/users/{user%2Did}/authentication/phoneMethods/{phoneAuthenticationMethod%2Did}/disableSmsSignIn";
    urlTplParams := make(map[string]string)
    for idx, item := range pathParameters {
        urlTplParams[idx] = item
    }
    m.pathParameters = urlTplParams
    m.requestAdapter = requestAdapter
    return m
}
// NewItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilder instantiates a new DisableSmsSignInRequestBuilder and sets the default values.
func NewItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilder(rawUrl string, requestAdapter i2ae4187f7daee263371cb1c977df639813ab50ffa529013b7437480d1ec0158f.RequestAdapter)(*ItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilder) {
    urlParams := make(map[string]string)
    urlParams["request-raw-url"] = rawUrl
    return NewItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilderInternal(urlParams, requestAdapter)
}
// Post disable SMS sign-in for an existing `mobile` phone number registered to a user. The number will no longer be available for SMS sign-in, which can prevent your user from signing in.
// [Find more info here]
// 
// [Find more info here]: https://docs.microsoft.com/graph/api/phoneauthenticationmethod-disablesmssignin?view=graph-rest-1.0
func (m *ItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilder) Post(ctx context.Context, requestConfiguration *ItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilderPostRequestConfiguration)(error) {
    requestInfo, err := m.ToPostRequestInformation(ctx, requestConfiguration);
    if err != nil {
        return err
    }
    errorMapping := i2ae4187f7daee263371cb1c977df639813ab50ffa529013b7437480d1ec0158f.ErrorMappings {
        "4XX": ia572726a95efa92ddd544552cd950653dc691023836923576b2f4bf716cf204a.CreateODataErrorFromDiscriminatorValue,
        "5XX": ia572726a95efa92ddd544552cd950653dc691023836923576b2f4bf716cf204a.CreateODataErrorFromDiscriminatorValue,
    }
    err = m.requestAdapter.SendNoContent(ctx, requestInfo, errorMapping)
    if err != nil {
        return err
    }
    return nil
}
// ToPostRequestInformation disable SMS sign-in for an existing `mobile` phone number registered to a user. The number will no longer be available for SMS sign-in, which can prevent your user from signing in.
func (m *ItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilder) ToPostRequestInformation(ctx context.Context, requestConfiguration *ItemAuthenticationPhoneMethodsItemDisableSmsSignInRequestBuilderPostRequestConfiguration)(*i2ae4187f7daee263371cb1c977df639813ab50ffa529013b7437480d1ec0158f.RequestInformation, error) {
    requestInfo := i2ae4187f7daee263371cb1c977df639813ab50ffa529013b7437480d1ec0158f.NewRequestInformation()
    requestInfo.UrlTemplate = m.urlTemplate
    requestInfo.PathParameters = m.pathParameters
    requestInfo.Method = i2ae4187f7daee263371cb1c977df639813ab50ffa529013b7437480d1ec0158f.POST
    if requestConfiguration != nil {
        requestInfo.Headers.AddAll(requestConfiguration.Headers)
        requestInfo.AddRequestOptions(requestConfiguration.Options)
    }
    return requestInfo, nil
}
