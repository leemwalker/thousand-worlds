import { json } from "@sveltejs/kit";
function GET() {
  return json({ status: "ok" });
}
export {
  GET
};
