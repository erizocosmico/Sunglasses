'use strict'

angular.module('sunglasses.controllers')
.controller('LandingController', ['$rootScope', '$scope', '$timeout', ($rootScope, $scope, $timeout) ->
    $rootScope.title = 'sunglasses'
    
    $scope.animateRedirect = (url) ->
        document.getElementById('landing').className += ' animated fadeOutUp'
        $timeout(() ->
            $rootScope.redirect(url)
        , 1000)
        
    introBg = document.getElementsByClassName('layer-intro')[0]
    footerBg = document.getElementsByClassName('layer-footer')[0]
    wrap = document.getElementById('landing')
    
    introBg.style.backgroundPositionY = '0px'
    footerBg.style.backgroundPositionY = '-100px'
    wrap.addEventListener('scroll', () ->
        if wrap.scrollTop < 250
            introBg.style.backgroundPositionY = '-' + (wrap.scrollTop) + 'px'
            footerBg.style.backgroundPositionY = '-100px'

        if wrap.scrollTop > (footerBg.offsetTop - 300) and wrap.scrollTop < footerBg.offsetTop
            footerBg.style.backgroundPositionY = '-' + (wrap.scrollTop - footerBg.offsetTop + 420) + 'px'
            introBg.style.backgroundPositionY = '0px'
    )
])