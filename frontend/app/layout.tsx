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
      <body>
        <AuthLoader />
        <Header />
        {children}
      </body>
    </html>
  );
}
