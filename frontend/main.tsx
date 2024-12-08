import * as vlens from "vlens";
import * as server from "@app/server"

async function main() {
    vlens.initRoutes([
        vlens.routeHandler("/users", () => import("@app/users")),
        vlens.routeHandler("/", () => import("@app/home")),
    ]);
}

main();

(window as any).server = server
