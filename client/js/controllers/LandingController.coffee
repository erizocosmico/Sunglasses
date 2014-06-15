'use strict'

angular.module('sunglasses.controllers')
.controller('LandingController', ['$rootScope', '$scope', '$timeout', ($rootScope, $scope, $timeout) ->
    $rootScope.title = 'sunglasses'
    
    $scope.animateRedirect = (url) ->
        document.getElementById('landing').className += ' animated fadeOutUp'
        $timeout(() ->
            $rootScope.redirect(url)
        , 1000)
])