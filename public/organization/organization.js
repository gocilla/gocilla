'use strict';

angular.module('organization', ['ngResource', 'uiSwitch'])

  .factory('OrganizationsService', ['$resource', '$cacheFactory', function($resource, $cacheFactory) {
    var organizationsCache = $cacheFactory('organizationsCache');
    return $resource('/api/organizations', {}, {query: {cache: organizationsCache, isArray: true}});
  }])

  .factory('RepositoryHookService', ['$resource', function($resource) {
    return $resource('/api/organizations/:orgId/repositories/:repoId/hook', {orgId: '@orgId', repoId: '@repoId'});
  }])

  .controller('OrganizationController', ['$scope', '$routeParams', '$cacheFactory',
        'OrganizationsService', 'RepositoryHookService',
        function($scope, $routeParams, $cacheFactory, OrganizationsService, RepositoryHookService) {
    $scope.orgId = $routeParams.orgId;
    $scope.organizations = OrganizationsService.query();
    $scope.switchRepo = function(repository) {
      if (repository.hooked) {
        RepositoryHookService.save({}, {orgId: $scope.orgId, repoId: repository.name});
      } else {
        RepositoryHookService.delete({}, {orgId: $scope.orgId, repoId: repository.name});
      }
      $cacheFactory.get('organizationsCache').removeAll();
    };
  }]);
