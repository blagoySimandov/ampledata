import { Navbar } from "./navbar";
import { Container } from "./container";

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <Container>
      <Navbar />
      {children}
    </Container>
  );
}
