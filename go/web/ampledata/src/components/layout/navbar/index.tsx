import {
  NavbarMain,
  NavbarContent,
  NavbarBrand,
  NavbarLogo,
  NavbarActions,
} from "./navbar";
import { Container } from "@/components/layout/container";
import { Button } from "@/components/ui/button";
import {
  NavigationMenu,
  NavigationMenuList,
  NavigationMenuItem,
  NavigationMenuLink,
} from "@/components/ui/navigation-menu";
import { useAuth } from "@workos-inc/authkit-react";
import { UserProfileMenu } from "@/components/auth";
import { UI_MESSAGES, APP_ROUTES } from "@/constants";
import logo from "@/assets/ampledata-high-resolution-logo-transparent.png";

export function Navbar() {
  const { user, signIn } = useAuth();

  return (
    <NavbarMain>
      <Container>
        <NavbarContent>
          <NavbarBrand>
            <NavbarLogo href={APP_ROUTES.HOME}>
              <img src={logo} alt="AmpleData" className="h-8" />
            </NavbarLogo>
            <NavigationMenu className="hidden md:flex">
              <NavigationMenuList>
                <NavigationMenuItem>
                  <NavigationMenuLink href={APP_ROUTES.ENRICHMENT} className="text-base">
                    Enrichment
                  </NavigationMenuLink>
                </NavigationMenuItem>
              </NavigationMenuList>
            </NavigationMenu>
          </NavbarBrand>

          <NavbarActions>
            {user ? (
              <UserProfileMenu />
            ) : (
              <Button onClick={() => signIn()}>
                {UI_MESSAGES.SIGN_IN}
              </Button>
            )}
          </NavbarActions>
        </NavbarContent>
      </Container>
    </NavbarMain>
  );
}
