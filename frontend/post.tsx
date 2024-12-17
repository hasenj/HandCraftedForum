import * as preact from "preact"
import * as vlens from "vlens";
import * as server from "@app/server";

export const postHandler = {
    fetch: fetchUserPosts,
    view: viewUserPosts
}

export const searchHandler = {
    fetch: fetchByHashtag,
    view: viewByHashtag,
}

export async function fetchUserPosts(route: string, prefix: string) {
    const params = vlens.urlParams(route);
    const userId = vlens.intParam(params, "user_id", 0);
    return server.QueryPosts({ Query: 'u:' + userId, Cursor: "" })
}

export async function fetchByHashtag(route: string, prefix: string) {
    const params = vlens.urlParams(route);
    const hashtag = params.get("hashtag") ?? "";
    return server.QueryPosts({ Query: 't:' + hashtag, Cursor: "" });
}

type Form = {
    content: string
    posts: server.Post[]
    nextQuery: server.PostsQuery
    sending: boolean
    error: string
}

const useForm = vlens.declareHook(
    (data: server.PostsResponse): Form => ({
        content: "",
        posts: data.Posts.slice(),
        nextQuery: data.NextParams,
        sending: false,
        error: "",
    })
)

async function fetchMorePosts(form: Form) {
    form.sending = true
    let [resp, err] = await server.QueryPosts(form.nextQuery)
    form.sending = false
    vlens.scheduleRedraw()
    if (resp) {
        form.posts.push(...resp.Posts)
        form.nextQuery = resp.NextParams
    } else {
        form.error = err
    }
}

const clsPosts = vlens.cssClass("posts", {
    display: "flex",
    flexDirection: "column",
    gap: "10px",
    width: "500px",
    padding: "10px 20px",
})

const clsPost = vlens.cssClass("post", {
    display: "flex",
    flexDirection: "column",
    gap: "4px",
    border: "1px solid silver",
    padding: "10px 20px",
    borderRadius: "8px",
    background: "#fefdff"
})

const clsTags = vlens.cssClass("tags", {
    display: "flex",
    gap: "4px",
    "a": {
        border: "1px solid silver",
        borderRadius: "4px",
        padding: "2px 4px",
        color: "#44a",
        background: "GhostWhite"
    }
})

const clsPostForm = vlens.cssClass("form", {
    display: "flex",
    flexDirection: "column",
    gap: "4px",
    alignItems: "flex-start",
})

const clsPostBody = vlens.cssClass("body", {
    margin: "10px 0",
    padding: "0",
})

const clsTimestamp = vlens.cssClass("timestamp", {
    fontSize: "12px",
    color: "gray",
})

const clsPostHeader = vlens.cssClass("header", {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
})

const clsPostAuthor = vlens.cssClass("author", {
    fontSize: "12px",
    color: "gray",
})

export function viewUserPosts(route: string, prefix: string, data: server.PostsResponse): preact.ComponentChild {
    const params = vlens.urlParams(route);
    const userId = vlens.intParam(params, "user_id", 0);
    const form = useForm(data);

    return <div>
        <form class={clsPostForm} onSubmit={vlens.cachePartial(onPostSubmit, userId, form)}>
            <textarea
                placeholder="Say something ..." rows={8} cols={80}
                {...vlens.attrsBindInput(vlens.ref(form, "content"))}
            />
            <button>Create</button>
        </form>
        {form.error && <div>Error: {form.error}</div>}
        {viewPosts(form)}
    </div>
}

async function onPostSubmit(userId: number, form: Form, event: Event) {
    event.preventDefault();
    form.sending = true;
    let [resp, err] = await server.CreatePost({ UserId: userId, Content: form.content });
    form.sending = false;
    vlens.scheduleRedraw();
    if (resp) {
        form.content = ""
        form.posts.unshift(resp)
    } else {
        form.error = err
    }
}

function viewPosts(form: Form) {
    return <div>
        <div class={clsPosts}>
            {form.posts.map(post => <div key={post.Id}>
                <div class={clsPost}>
                    <div class={clsPostHeader}>
                        <div class={clsPostAuthor}>@{post.UserId}</div>
                        <div class={clsTimestamp}>{postTimestamp(post.CreatedAt)}</div>
                    </div>
                    <div class={clsPostBody}>{post.Content}</div>
                    <div class={clsTags}>
                        {post.Tags.map(tag =>
                            <div>
                                <a key={tag} href={`/search?hashtag=${tag}`}>#{tag}</a>
                            </div>
                        )}
                    </div>
                </div>
            </div>)}
        </div>
        {form.nextQuery.Cursor && <button disabled={form.sending}
            onClick={vlens.cachePartial(fetchMorePosts, form)}>More</button>}
    </div>
}

// input string is json, output string is human readable
function postTimestamp(ts: string): string {
    const date = new Date(ts)
    // formatted as yyyy-mm-dd hh:mm
    const year = date.getFullYear();
    const month = date.getMonth() + 1;
    const day = date.getDate();
    const hour = date.getHours();
    const minute = date.getMinutes();

    const pad = (n: number) => n.toString().padStart(2, "0")
    return `${year}-${pad(month)}-${pad(day)} ${pad(hour)}:${pad(minute)}`;
}

// ---


export function viewByHashtag(route: string, prefix: string, data: server.PostsResponse) {
    const form = useForm(data);
    return viewPosts(form)
}
