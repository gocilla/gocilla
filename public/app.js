'use strict';

angular.module('gocilla', ['ngResource', 'ngRoute', 'profile', 'organization', 'repository'])

  .config(function ($routeProvider, $locationProvider) {
    $routeProvider
      .when('/', {
        templateUrl: '/public/organization/organization.html',
        controller: 'OrganizationController'
      })
      .when('/organizations', {
        templateUrl: '/public/organization/organization.html',
        controller: 'OrganizationController'
      })
      .when('/organizations/:orgId', {
        templateUrl: '/public/organization/organization.html',
        controller: 'OrganizationController'
      })
      .when('/organizations/:orgId/repositories/:repoId', {
        templateUrl: '/public/repository/repository.html',
        controller: 'RepositoryController'
      })
      .when('/organizations/:orgId/repositories/:repoId/builds/:buildId', {
        templateUrl: '/public/repository/build.html',
        controller: 'RepositoryController'
      })
      .when('/organizations/:orgId/repositories/:repoId/hook', {
        templateUrl: '/public/organization/organization.html',
        controller: 'OrganizationController'
      })
      .otherwise({
        redirectTo: '/'
      });
    $routeProvider.caseInsensitiveMatch = true;
    $locationProvider.html5Mode(true);
  })
  .filter('moment', function () {
    return function (input, momentFn /*, param1, param2, ...param n */) {
      var args = Array.prototype.slice.call(arguments, 2),
          momentObj = moment(input);
      return momentObj[momentFn].apply(momentObj, args);
    };
  })
  .filter('duration', function () {
    return function (start, end) {
      var m1 = moment(start);
      var m2 = moment(end);
      return moment.duration(m1.diff(m2)).humanize();
    };
  })
  .directive('status', function() {
    return {
      restrict: 'E',
      scope: {
        status: '='
      },
      templateUrl: '/public/templates/status.html'
    }
  })
  .directive('navrepo', function() {
    return {
      restrict: 'E',
      templateUrl: '/public/templates/navrepo.html'
    }
  })
  .directive('navigation', function() {
    return {
      restrict: 'AE',
      replace: 'true',
      templateUrl: '/public/templates/navigation.html'
    }
  });
