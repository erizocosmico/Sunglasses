'use strict'

angular.module('sunglasses')
.directive('footerBar', () ->
    restrict: 'E',
    templateUrl: 'templates/footer-bar.html',
    controller: [
        '$scope',
        '$rootScope',
        '$translate',
        'api',
        ($scope, $rootScope, $translate,  api) ->
            $scope.preferredLang = $rootScope.userData.preferred_lang || 'en'
            $scope.languages =
                es: 'Español',
                en: 'English'

            $scope.changeLanguage = () ->
                api(
                    '/api/users/change_lang',
                    'PUT',
                    lang: $scope.preferredLang,
                    (resp) ->
                        $translate.use($scope.preferredLang)
                )
    ]
)