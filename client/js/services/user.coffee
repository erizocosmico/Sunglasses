'use strict'

angular.module('sunglasses.services')
# gives functionality to interact with user entities
.factory('user', () ->
    # retrieves the user avatar thumbnail (it can be public or private)
    getAvatarThumb: (user) ->
        if !user.avatar_thumbnail?
            return user.avatar_thumbnail
        else
            return user.public_avatar_thumbnail
    # retrieves the user avatar (it can be public or private)   
    , getAvatar: (post) ->
        if !user.avatar?
            return user.avatar
        else
            return user.public_avatar   
    # retrieves the username (it can be public or private)
    , getUsername: (user) ->
        if !user.private_name?
            return user.private_name
        else if !user.public_name?
            return user.public_name
        else
            return user.username
)