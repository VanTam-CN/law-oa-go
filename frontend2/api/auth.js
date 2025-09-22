import request from '../utils/request';

export async function login(data) {
  return request({
    url: '/api/v1/auth/login',
    method: 'post',
    data
  });
}

export async function register(data) {
  return request({
    url: '/api/v1/auth/register',
    method: 'post',
    data
  });
}

export async function refreshToken(data) {
  return request({
    url: '/api/v1/auth/refresh',
    method: 'post',
    data
  });
}

export async function logout() {
  return request({
    url: '/api/v1/users/logout',
    method: 'post'
  });
}

export async function getCurrentUser() {
  return request({
    url: '/api/v1/users/profile',
    method: 'get'
  });
}

export async function updateProfile(data) {
  return request({
    url: '/api/v1/users/profile',
    method: 'put',
    data
  });
}

export async function changePassword(data) {
  return request({
    url: '/api/v1/users/change-password',
    method: 'post',
    data
  });
}