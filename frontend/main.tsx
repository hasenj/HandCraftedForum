import * as vlens from "vlens";

async function main() {
    vlens.initRoutes([
        vlens.routeHandler("/", () => import("@app/home")),
    ]);
}

main();
