import "./globals.css";
import AuthLoader from "./components/AuthLoader";
import Header from "./components/Header";

export const metadata = {
  title: "Sol Coffee System",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ja">
      <body className="min-h-screen bg-zinc-50 text-zinc-900 antialiased">
        <AuthLoader />
        <Header />
        <div className="pb-12">{children}</div>
      </body>
    </html>
  );
}
