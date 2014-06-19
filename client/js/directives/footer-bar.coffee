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
            $scope.preferredLang = $rootScope?.userData?.preferred_lang || window.localStorage?.getItem('preferred_lang') || 'en'
            $scope.languages =
                es: 'Español',
                en: 'English'

            $scope.changeLanguage = () ->
                if userData?
                    api(
                        '/api/users/change_lang',
                        'PUT',
                        lang: $scope.preferredLang,
                        (resp) ->
                            $scope.$apply(() ->
                                $translate.use($scope.preferredLang)
                            )
                    )
                else
                    $translate.use($scope.preferredLang)
                    window.localStorage?.setItem('preferred_lang', $scope.preferredLang)
    ]
)