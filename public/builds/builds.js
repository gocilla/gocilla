'use strict';

angular.module('builds', ['ngResource'])

  .factory('BuildsService', ['$resource', '$cacheFactory', function($resource, $cacheFactory) {
    var buildsCache = $cacheFactory('buildsCache');
    return $resource('/api/builds', {}, {query: {cache: buildsCache, isArray: true}});
  }])

  .controller('BuildsController', ['$scope', '$routeParams', '$cacheFactory', 'BuildsService',
        function($scope, $routeParams, $cacheFactory, BuildsService) {
    $scope.buildId = $routeParams.buildId;
    $scope.builds = BuildsService.query();
  }]);
