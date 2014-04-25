'use strict'

angular.module('mask.services')
# api is a shortcut to perform api calls
.factory('api', () ->
    (url, method, params, success, error) ->
        timestamp = new Date().getTime()
        params.signature = md5(url + csrfToken + timestamp)
        params.timestamp = timestamp
        
        $.ajax(
            url: url,
            method: method,
            dataType: 'json',
            data: params
            success: (resp) ->
                success(resp)
            error: (resp) ->
                if error?
                    error(resp)
                else
                    # TODO: Default error handler
                    console.log('Error')
        )
)