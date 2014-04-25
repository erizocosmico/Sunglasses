'use strict'

angular.module('mask.services')
.factory('apiSignature', ['$location', ($location) ->
    timestamp = new Date().getTime()
    signature: md5($location.path() + csrfToken + timestamp),
    timestamp: timestamp
])