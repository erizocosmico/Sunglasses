'use strict'

angular.module('sunglasses.controllers')
.controller('LoginController', [
    '$scope',
    '$rootScope',
    'api',
    ($scope, $rootScope, api) ->
        $rootScope.title = 'login'

        # username and password
        $scope.data =
            username: ''
            password: ''
        
        # data has been submitted
        submitted = false
        
        # loginClick performs an api call to login the user
        # if successful the user will be redirected to the home page
        $scope.loginClick = () ->
            valid = true
            if $scope.data.password.length < 6
                valid = false
                $rootScope.displayError('login-password-error')
                
            if not /^[a-zA-Z_0-9]{2,30}$/.test($scope.data.username)
                valid = false
                $rootScope.displayError('login-username-error')
            
            if valid and not submitted
                submitted = true
                api('/api/auth/login',
                    'POST',
                    $scope.data,
                    (resp) ->
                        $rootScope.fullRefresh()
                    (resp) ->
                        submitted = false
                        $rootScope.displayError('login-invalid-error')
                )
])
