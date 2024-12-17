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
    CreatedAt: string
    Content: string
    Tags: string[]
}

export interface PostsQuery {
    Query: string
    Cursor: string
}

export interface PostsResponse {
    Posts: Post[]
    NextParams: PostsQuery
}

export interface User {
    Id: number
    Username: string
    Email: string
    IsAdmin: boolean
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

export async function QueryPosts(data: PostsQuery): Promise<rpc.Response<PostsResponse>> {
    return await rpc.call<PostsResponse>('QueryPosts', JSON.stringify(data));
}

