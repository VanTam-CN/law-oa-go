import request from '../utils/request';

export async function getDepartmentList(params) {
  return request({
    url: '/lf/dept/list',
    method: 'get',
    params
  });
}

export async function getDepartmentDetail(deptId) {
  return request({
    url: `/lf/dept/${deptId}`,
    method: 'get'
  });
}

export async function createDepartment(data) {
  return request({
    url: '/lf/dept',
    method: 'post',
    data
  });
}

export async function updateDepartment(data) {
  return request({
    url: '/lf/dept',
    method: 'put',
    data
  });
}

export async function deleteDepartment(deptId) {
  return request({
    url: `/lf/dept/${deptId}`,
    method: 'delete'
  });
}

export async function getDepartmentTree(params) {
  return request({
    url: '/lf/dept/treeselect',
    method: 'get',
    params
  });
}

export async function getDepartmentExcludeChild(deptId) {
  return request({
    url: `/lf/dept/list/exclude/${deptId}`,
    method: 'get'
  });
}

export async function getRoleDepartmentTree(roleId) {
  return request({
    url: `/lf/dept/roleDeptTreeselect/${roleId}`,
    method: 'get'
  });
}