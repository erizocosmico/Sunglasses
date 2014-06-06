'use strict'

angular.module('sunglasses.controllers')
.controller('SettingsController', [
    '$scope',
    '$rootScope',
    '$translate',
    'api',
    ($scope, $rootScope, $translate, api) ->
        $rootScope.title = 'settings'
        
        # Settings
        $scope.settings = userData.settings
        # Info
        $scope.info = userData.info
        # Profile data
        $scope.profileData = 
            public_name: $rootScope.userData.public_name
            private_name: $rootScope.userData.private_name
        # Password change model
        $scope.passwordChange = 
            password: ''
            password_repeat: ''
            current_password: ''
        # Avatars
        $scope.avatars = 
            public: null
            private: null
        
        # Active section
        $scope.activeSection = 'account_details'
        $scope.setActiveSection = (section) ->
            $scope.activeSection = section
        
        for t in ['private', 'public']
            ((type) ->
                document.getElementById(type + '-avatar').addEventListener('change', (e) ->
                    $scope.avatars[type] = e.target.files[0]
                    $scope.uploadAvatar(type)
                )
            )(t)
            
        $scope.handleUpload = (type) ->
            if ['private', 'public'].indexOf(type) == -1 then return
            
            e = document.createEvent('Event')
            e.initEvent('click', true, true)
            document.getElementById(type + '-avatar').dispatchEvent(e)
            
        $scope.uploadAvatar = (type) ->
            $translate('confirm_avatar_upload')
            .then((text) ->
                if confirm(text)
                    api(
                        '/api/account/update_picture',
                        'PUT',
                        account_picture: $scope.avatars[type],
                        picture_type: type,
                        (resp) ->
                            # TODO: backend should return thumbnail url
                            console.log(resp)
                        , (resp) ->
                            console.log(resp)
                    )
            )
            
        $scope.updatePassword = () ->
            api(
                '/api/account/password',
                'PUT',
                $scope.passwordChange,
                (resp) ->
                    console.log(resp)
                , (resp) ->
                    console.log(resp)
            )
            
        $scope.updateData = () ->
            api(
                '/api/account/data',
                'PUT',
                $scope.profileData,
                (resp) ->
                    console.log(resp)
                , (resp) ->
                    console.log(resp)
            )

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
