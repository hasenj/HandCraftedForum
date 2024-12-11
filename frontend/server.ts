import * as rpc from "vlens/rpc"

export interface AddUserRequest {
    Username: string
    Email: string
    Password: string
}

export interface UserListResponse {
    Users: User[]
}

export interface Empty {
}

export interface CreatePostReq {
    UserId: number
    Content: string
}

export interface Post {
    Id: number
    UserId: number
    CreatedAt: Time
    Content: string
}

export interface ByUserReq {
    UserId: number
}

export interface Posts {
    Posts: Post[]
}

export interface ByHashtagReq {
    Hashtag: string
}

export interface User {
    Id: number
    Username: string
    Email: string
    IsAdmin: boolean
}

export interface Time {
    wall: number
    ext: number
    loc: Location | null
}

export interface Location {
    name: string
    zone: zone[]
    tx: zoneTrans[]
    extend: string
    cacheStart: number
    cacheEnd: number
    cacheZone: zone | null
}

export interface zone {
    name: string
    offset: number
    isDST: boolean
}

export interface zoneTrans {
    when: number
    index: number
    isstd: boolean
    isutc: boolean
}

export async function AddUser(data: AddUserRequest): Promise<rpc.Response<UserListResponse>> {
    return await rpc.call<UserListResponse>('AddUser', JSON.stringify(data));
}

export async function ListUsers(data: Empty): Promise<rpc.Response<UserListResponse>> {
    return await rpc.call<UserListResponse>('ListUsers', JSON.stringify(data));
}

export async function CreatePost(data: CreatePostReq): Promise<rpc.Response<Post>> {
    return await rpc.call<Post>('CreatePost', JSON.stringify(data));
}

export async function PostsByUser(data: ByUserReq): Promise<rpc.Response<Posts>> {
    return await rpc.call<Posts>('PostsByUser', JSON.stringify(data));
}

export async function PostsByHashtag(data: ByHashtagReq): Promise<rpc.Response<Posts>> {
    return await rpc.call<Posts>('PostsByHashtag', JSON.stringify(data));
}

