import request from '../utils/request';

export async function getRoleList(params) {
  return request({
    url: '/lf/role/list',
    method: 'get',
    params
  });
}

export async function getRoleDetail(roleId) {
  return request({
    url: `/lf/role/${roleId}`,
    method: 'get'
  });
}

export async function createRole(data) {
  return request({
    url: '/lf/role',
    method: 'post',
    data
  });
}

export async function updateRole(data) {
  return request({
    url: '/lf/role',
    method: 'put',
    data
  });
}

export async function deleteRole(roleIds) {
  return request({
    url: `/lf/role/${roleIds}`,
    method: 'delete'
  });
}

export async function changeRoleStatus(data) {
  return request({
    url: '/lf/role/changeStatus',
    method: 'put',
    data
  });
}

export async function getRoleOptions() {
  return request({
    url: '/lf/role/optionselect',
    method: 'get'
  });
}

export async function getAllocatedUserList(params) {
  return request({
    url: '/lf/role/authUser/allocatedList',
    method: 'get',
    params
  });
}

export async function getUnallocatedUserList(params) {
  return request({
    url: '/lf/role/authUser/unallocatedList',
    method: 'get',
    params
  });
}

export async function cancelAuthUser(data) {
  return request({
    url: '/lf/role/authUser/cancel',
    method: 'put',
    data
  });
}

export async function cancelAuthUserAll(roleId, userIds) {
  return request({
    url: `/lf/role/authUser/cancelAll?roleId=${roleId}&userIds=${userIds.join(',')}`,
    method: 'put'
  });
}

export async function selectAuthUserAll(roleId, userIds) {
  return request({
    url: `/lf/role/authUser/selectAll?roleId=${roleId}&userIds=${userIds.join(',')}`,
    method: 'put'
  });
}