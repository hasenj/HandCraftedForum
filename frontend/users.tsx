import * as preact from "preact"
import * as vlens from "vlens";
import * as server from "@app/server";

export async function fetch(route: string, prefix: string) {
    return server.ListUsers({})
}

type Form = {
    users: server.User[]
    name: string
    email: string
    password: string
    error: string
}

const useForm = vlens.declareHook((data: server.UserListResponse): Form => ({
    users: data.Users, name: "", email: "", password: "", error: ""
}))

export function view(route: string, prefix: string, data: server.UserListResponse): preact.ComponentChild {
    let form = useForm(data)
    return <div>
        <h3>Users</h3>
        {form.users.map(user => <div key={user.Id}><a href={`/post?user_id=${user.Id}`}>{user.Username}</a></div>)}
        <h3>Add User</h3>
        <form onSubmit={vlens.cachePartial(onAddUserClicked, form)}>
            <div>
                <label>Username:
                    <input type="text" {...vlens.attrsBindInput(vlens.ref(form, "name"))} />
                </label>
            </div>
            <div>

                <label>Email:
                    <input type="text" {...vlens.attrsBindInput(vlens.ref(form, "email"))} />
                </label>
            </div>
            <div>

                <label>Password:
                    <input type="password" {...vlens.attrsBindInput(vlens.ref(form, "password"))} />
                </label>
            </div>
            <button onClick={vlens.cachePartial(onAddUserClicked, form)}>Add</button>
        </form>
    </div>
}

async function onAddUserClicked(form: Form, event: Event) {
    event.preventDefault()

    let [resp, err] = await server.AddUser({
        Username: form.name, Email: form.email, Password: form.password
    })
    if (resp) {
        form.name = ""
        form.password = ""
        form.email = ""
        form.users = resp.Users
        form.error = ""
    } else {
        form.error = err
    }
    vlens.scheduleRedraw()
}
