'use strict'

angular.module('sunglasses.services')
# gives functionality to interact with photos
.factory('photo', ['$rootScope', ($rootScope) ->
    $theater = 
        open: (image, caption) ->
            theater = $('#photo-theater').get(0)
            photoHolder = $('#photo-theater-photo').get(0).parentNode
            photoHolder.innerHTML = ''
            photo = document.createElement('img')
            photo.id = 'photo-theater-photo'
            photoHolder.appendChild(photo)

            photo.alt = caption
            photo.src = image
        
            h = window.innerHeight
            w = window.innerWidth
        
            photo.style.marginTop = '0px'
            if photo.width > w then photo.width = w

            if photo.height > h
                photo.height = h
            else
                photo.style.marginTop = ((h - photo.height) / 2) + 'px'

            $rootScope.animateElem(theater, 'fadeInDown')  
            return
        , close: () ->
            theater = $('#photo-theater').get(0)
            $rootScope.animateElem(theater, 'fadeOutUp', () ->
                theater.className = 'hidden'
            )
            return
    
    openTheater: (image, caption) ->
        $theater.open(image, caption)
        return
    , dismissTheater: () ->
        $theater.close()
        return
])