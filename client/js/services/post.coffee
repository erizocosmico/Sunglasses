'use strict'

angular.module('sunglasses.services')
# gives functionality to interact with post entities
.factory('post', () ->
    # tells if the post is a certain kind of post
    isKindOf: (post, kind) ->
        valid = true
        switch kind
            when 'video'
                valid = post.video_id? and post.video_service?
            when 'link'
                valid = post.link_url?
            when 'photo'
                valid = post.photo_url? and post.thumbnail?
        return valid
)