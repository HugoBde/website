use std::fs;

use axum::Router;
use axum::http::StatusCode;
use axum::response::Html;
use axum::routing::{get, get_service};
use axum_extra::extract::CookieJar;
use tower_http::services::ServeDir;

#[tokio::main]
async fn main() {
  let app = Router::new()
    .route("/", get(get_home))
    .nest_service("/dist", get_service(ServeDir::new("dist")));

  let listener = tokio::net::TcpListener::bind("0.0.0.0:8443").await.unwrap();
  axum::serve(listener, app).await.unwrap();
}

async fn get_home(jar: CookieJar) -> Result<Html<String>, StatusCode> {
  if let Some(c) = jar.get("version") {
    match c.value() {
      "base" => serve_file("dist/base/index.html"),
      "shell" => serve_file("dist/shell/index.html"),
      _ => Err(StatusCode::BAD_REQUEST),
    }
  } else {
    serve_file("dist/base/index.html")
  }
}

fn serve_file(file: &'static str) -> Result<Html<String>, StatusCode> {
  match fs::read_to_string(file) {
    Ok(content) => Ok(Html(content)),
    Err(err) => {
      eprintln!("Error reading HTML file {}: {}", file, err);
      Err(axum::http::StatusCode::INTERNAL_SERVER_ERROR)
    }
  }
}
