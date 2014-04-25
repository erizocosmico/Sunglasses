'use strict'

angular.module('mask.services')
# apiSignature returns an object with the request signature and timestamp
.factory('apiSignature', ['$location', ($location) ->
    timestamp = new Date().getTime()
    signature: md5($location.path() + csrfToken + timestamp),
    timestamp: timestamp
])

# api is a shortcut to perform api calls
.factory('api', ['$http', 'apiSignature', ($http, apiSignature) ->
    (url, method, params, success, error) ->
        angular.extend(params, apiSignature)
        $http(angular.extend(
            url: url,
            method: method,
            if method == 'POST' then {data: params} else {params: params}
        )).success((resp) ->
            success(resp)
        ).error((resp) ->
            if error?
                error(resp)
            else
                # TODO: Default error handler
                console.log('Error')
        )
])