'use strict'

angular.module('sunglasses')
# displays a photo theater
.directive('photoTheater', () ->
    restrict: 'E',
    replace: true,
    template: '<div id="photo-theater" class="hidden">
                    <button class="btn-close-theater" ng-click="photoService.dismissTheater()">&times;</button>
                    <div class="photo-holder">
                        <div><img id="photo-theater-photo"></div>
                    </div>
                </div>'
)