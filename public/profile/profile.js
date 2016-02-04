'use strict';

angular.module('profile', ['ngResource'])

  .factory('ProfileService', ['$resource', '$cacheFactory', function($resource, $cacheFactory) {
    var profileCache = $cacheFactory('profileCache');
    return $resource('/api/profile', {}, {query: {cache: profileCache}});
  }])

  .factory('LogoutService', ['$resource', function($resource) {
    return $resource('/logout');
  }])

  .controller('ProfileController', ['$scope', '$window','ProfileService', 'LogoutService',
        function($scope, $window, ProfileService, LogoutService) {
    $scope.profile = ProfileService.get();
    $scope.logout = function() {
      var logout = LogoutService.get();
      $window.location.href = '/';
    };

    $scope.login = function() {
      $window.location.href = '/login';
    };
  }]);
