'use strict'

angular.module('sunglasses.controllers')
.controller('SettingsController', [
    '$scope',
    '$rootScope',
    'api',
    ($scope, $rootScope, api) ->
        $rootScope.title = 'settings'
        
        # Settings
        $scope.settings = userData.settings
        # Info
        $scope.info = userData.info
        
        # Active section
        $scope.activeSection = 'account_details'
        $scope.setActiveSection = (section) ->
            $scope.activeSection = section

        # update user settings
        $scope.updateSettings = () ->
            data = $scope.settings
            for pType in ['status', 'photo', 'video', 'link', 'album']
                data['privacy_'+pType+'_type'] = $scope.settings['default_'+pType+'_privacy'].privacy_type

                if 'privacy_users' in $scope.settings['default_'+pType+'_privacy']
                    data['privacy_'+pType+'_users'] = $scope.settings['default_'+pType+'_privacy'].privacy_users

            api(
                '/api/account/settings',
                'PUT',
                data,
                (resp) ->
                    console.log resp

                    # Update user data
                    userData.settings = $scope.settings
                , (resp) ->
                    console.log resp
            )
            
        # update user info
        $scope.updateInfo = () ->
            api(
                '/api/account/info',
                'PUT',
                $scope.info,
                (resp) ->
                    console.log resp

                    # Update user data
                    userData.info = $scope.info
                , (resp) ->
                    console.log resp
            )

        # Workaround for Semantic's problems with Angular
        $scope.toggle = (key) ->
            $scope.settings[key] = !$scope.settings[key]
            
        (() ->
            $('.ui.checkbox').checkbox()

            # value needs to be applied, another semantic ui workaround
            for pType in ['status', 'photo', 'video', 'link']
                ((t) ->
                    $('#selector_'+t+'_privacy')
                    .dropdown('set active')
                    .dropdown('set value', $scope.settings['default_'+t+'_privacy'].privacy_type)
                    .dropdown(
                        onChange: (val) ->
                            $scope.$apply(() ->
                                $scope.settings['default_'+t+'_privacy'].privacy_type = val
                            )
                    )
                )(pType)

            for key in ['gender', 'status']
                ((k) ->
                    $('#selector-' + k)
                    .dropdown('set value', $scope.info[k])
                    .dropdown(
                        onChange: (val) ->
                            $scope.$apply(() ->
                                $scope.info[k] = val
                            )
                    )
                )(key)
                
            # Fixed menu
            menu = $('aside').get(0)
            $('#app').scroll(() ->
                if $(this).scrollTop() > 30
                    width = $(menu).width()
                    menu.style.position = 'fixed'
                    menu.style.width = width + 'px'
                else
                    menu.style.position = 'static'
            )
        )()
])
