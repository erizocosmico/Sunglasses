'use strict'

angular.module('sunglasses')
# renders a video element from vimeo or youtube
.directive('sunVideo', () ->
    getVideoWidget = (service, id) ->
        if Number(service) == 1
            return '<iframe width="500" height="281" 
            src="//www.youtube-nocookie.com/embed/'+id+'?rel=0" 
            frameborder="0" allowfullscreen></iframe>'
        else
            return '<iframe src="//player.vimeo.com/video/'+id+'?color=1b71ad" 
            width="500" height="281" frameborder="0" 
            webkitallowfullscreen mozallowfullscreen allowfullscreen></iframe>'
    
    restrict: 'E',
    replace: true,
    template: '<div compile="widget"></div>',
    link: (scope, elem, attrs) ->
        scope.widget = getVideoWidget(attrs.service, attrs.videoId)
)