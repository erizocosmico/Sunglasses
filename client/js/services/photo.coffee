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
            
            h = window.innerHeight
            w = window.innerWidth
            
            $("<img/>").attr("src", image).load(() ->
                photo.alt = caption
                photo.src = image
                
                if this.height > h
                    this.width = (this.width * h) / this.height
                    this.height = h
                    
                if this.width > w
                    this.height = (this.height * w) / this.width
                    this.width = w
                    
                photo.width = this.width
                photo.height = this.height
                    
                if this.height < h
                    photo.style.marginTop = ((h - photo.height) / 2) + 'px'
                else
                    photo.style.marginTop = '0px'
            )

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