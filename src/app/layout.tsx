import type { Metadata } from "next";
import { DM_Sans } from "next/font/google";
import "./globals.css";

// Self-hosted by next/font: no runtime request to Google, no layout shift.
const dmSans = DM_Sans({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-dm-sans",
  weight: ["400", "500", "700", "800", "900"]
});

export const metadata: Metadata = {
  title: "BRAG",
  description: "BNI Grow annual member challenge platform"
};

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id" className={dmSans.variable}>
      <body className="font-sans antialiased">{children}</body>
    </html>
  );
}
