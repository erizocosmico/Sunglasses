'use strict'

angular.module('sunglasses.services')
# api is a shortcut to perform api calls
.factory('api', () ->
    (url, method, params, success, error) ->
        timestamp = new Date().getTime()
        params.signature = md5(url + csrfToken + timestamp)
        params.timestamp = timestamp
        
        if method != 'GET'
            data = new FormData()
            for key, val of params
                data.append(key, val)
        else
            data = params

        $.ajax(
            url: url,
            method: method,
            cache: false,
            dataType: 'json',
            processData: method == 'GET',
            contentType: false,
            data: data,
            success: (resp) ->
                success(resp)
            , error: (resp) ->
                if error?
                    error(resp)
                else
                    # TODO: Default error handler
                    console.log('Error')
        )
)